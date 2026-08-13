# Custom Actions

Custom actions let a **driver** add its own actions to any built-in object, on top of
the predefined ones the SDK ships with — without forking the SDK or implementing a new
object type.

Use them when a device supports an operation that the built-in action set doesn't cover:
reboot, calibrate, blink an LED, run a self-test, sync a clock, trigger a relay pulse, etc.

- **Available since:** SDK `v0.7.78`
- **Source:** [`pkg/objects/custom_actions.go`](../pkg/objects/custom_actions.go),
  dispatch in [`pkg/objects/object_runner.go`](../pkg/objects/object_runner.go)
- **Works on:** every built-in object type — `switch`, `door`, `lock`, `sensor`, `speaker`,
  `microphone`, `video_channel`, `video_engine`, `alarm_panel`, `reader`, `access_controller`,
  `octopus`, `notifier`, `gps_tracker`, `person`, `relative_zone`, `relative_tracker`.

---

## Contents

1. [TL;DR](#tldr)
2. [Lifecycle end to end](#lifecycle-end-to-end)
3. [API reference](#api-reference)
4. [Routing: actions are dispatched by domain](#routing-actions-are-dispatched-by-domain)
5. [The payload contract](#the-payload-contract)
6. [Returning results and errors](#returning-results-and-errors)
7. [Concurrency and panics](#concurrency-and-panics)
8. [Using the controller from a handler](#using-the-controller-from-a-handler)
9. [Naming](#naming)
10. [Patterns](#patterns)
11. [Full example](#full-example)
12. [Testing](#testing)
13. [Troubleshooting](#troubleshooting)
14. [Limitations](#limitations)
15. [FAQ](#faq)

---

## TL;DR

```go
sw := objects.NewSwitchObject(params)

// 1. Register the action BEFORE registering the object with the runner.
err := sw.RegisterCustomAction("switch.action.blink", func(ctx objects.CustomActionContext) (map[string]string, error) {
    var p struct {
        Times int `json:"times"`
    }
    if err := json.Unmarshal(ctx.Payload, &p); err != nil {
        return nil, fmt.Errorf("invalid payload: %w", err)
    }
    if p.Times <= 0 {
        return nil, fmt.Errorf("times must be > 0")
    }
    // ... drive the device ...
    return map[string]string{"status": "blinked"}, nil
})
if err != nil {
    log.Fatal(err) // registration errors are programming errors — fail fast
}

// 2. Now register the object — the custom action is published to the platform here.
runner.RegisterObject(sw)
```

---

## Lifecycle end to end

Custom actions ride on the two contracts every object already exposes, so nothing in the
dispatch pipeline changes.

```
  BUILD TIME (your driver's code)
  ────────────────────────────────
  obj := objects.NewSwitchObject(...)
  obj.RegisterCustomAction("x.action.y", handler)   ──► stored in the object's registry


  REGISTRATION (client.RegisterObject → objectRunner.RegisterObject)
  ──────────────────────────────────────────────────────────────────
  1. controller.CreateObject(obj)                    POST /objects
  2. for a := range obj.GetAvailableActions():       ◄── built-ins + your custom actions
         controller.NewAction(a)                     POST /objects/actions
  3. objectsMap[obj.Domain] = append(..., obj)       ◄── indexed BY DOMAIN
  4. publish SUBSCRIBE_OBJECTS_COMMANDS_LISTENING
  5. obj.Setup(controller)                           ◄── ctx.Controller becomes usable HERE


  RUNTIME (platform triggers the action)
  ───────────────────────────────────────
  DriversHub ──REQUEST_ACTION_EXECUTION──► objectRunner.listenActions()
       │
       ├─ look up objectsMap[req.domain]
       ├─ filter by req.object_id  (EMPTY ⇒ every object in the domain)
       └─ go obj.RunAction(execID, action, payloadBytes)     ◄── one goroutine per object
                    │
                    ▼
             switch action {
               case <built-in>: ...          ◄── built-ins are matched FIRST
               default:
                 dispatchCustom(...)         ◄── your handler, or "action %s not found"
             }
                    │
       ┌────────────┴────────────┐
       │ err != nil              │ err == nil
       ▼                         ▼
  UpdateResultAttributes(    UpdateResultAttributes(
    execID,                    execID, resp)
    {"error": err.Error()})
       │                         │
       └────────► PUT /objects/actions/executions/{execID}
                  body: {"result": { ...map... }}
```

Two consequences worth internalizing:

- **Built-ins are matched before the registry**, so a custom action can never shadow a
  built-in one. A colliding name silently never reaches your handler.
- **`GetAvailableActions()` is read exactly once**, inside `RegisterObject`. Anything
  registered after that call still *dispatches* correctly, but is never advertised to
  the platform, so nothing in the UI can trigger it.

---

## API reference

```go
// Implemented by every built-in object.
type CustomActionRegistrar interface {
    RegisterCustomAction(action string, handler CustomActionHandler) error
}

// Your handler. The returned map becomes the action execution result reported
// back to the platform.
type CustomActionHandler func(ctx CustomActionContext) (map[string]string, error)

// Everything the handler receives.
type CustomActionContext struct {
    ExecutionID string            // action execution id from the platform
    Action      string            // the custom action name being invoked
    Payload     []byte            // raw JSON payload sent by DriversHub
    Object      RegistrableObject // the object the action was invoked on
    Controller  ObjectController  // nil until the object's Setup() has run
}
```

### `RegisterCustomAction(action string, handler CustomActionHandler) error`

Registers a handler under `action`. Returns an error when:

| Condition | Error |
|---|---|
| `action == ""` | `custom action name cannot be empty` |
| `handler == nil` | `custom action handler cannot be nil` |
| `action` already registered on this object | `custom action "<name>" already registered` |

All three are programming errors, not runtime conditions — check the error and fail fast
at startup rather than logging and continuing with a half-configured object.

The registry is safe for concurrent use (`sync.RWMutex`); registration and dispatch may
overlap without data races.

### What the platform learns about your action

`GetAvailableActions()` advertises each custom action as:

```go
type ObjectAction struct {
    Action string `json:"action"` // the name you registered
    Domain string `json:"domain"` // the object's domain
}
```

That is the **entire** contract published to DriversHub. There is no field for a payload
schema, parameter list, label, description, or icon — see
[Limitations](#limitations) for what that implies.

---

## Routing: actions are dispatched by domain

This is the single most common source of surprise, so it gets its own section.

The runner indexes registered objects **by `ObjectMetadata.Domain`**, not by object type
and not by object ID:

```go
o.objectsMap.Store(object.GetMetadata().Domain, append(existingObjects, object))
```

An incoming `REQUEST_ACTION_EXECUTION` carries `domain`, `action`, `id`, and an
**optional** `object_id` array. The runner then:

| `object_id` in the request | What runs |
|---|---|
| Non-empty | `RunAction` on each object in that domain whose `ObjectID` matches |
| Empty / absent | `RunAction` on **every object registered under that domain** |

Three practical consequences:

**1. Fan-out is real.** If you register 40 cameras under domain `"camera"` and the platform
sends a domain-wide execution, all 40 handlers run concurrently against 40 devices. Handlers
must tolerate being invoked in bulk.

**2. Objects in a domain that lack the handler report an error.** Suppose `door-1` registers
`door.action.pulse` and `door-2` (same domain) does not. A domain-wide execution calls both;
`door-2` falls through to `dispatchCustom`, finds nothing, and returns
`action door.action.pulse not found` — which the runner writes back as
`{"error": "action door.action.pulse not found"}`. Since every object shares one
`ExecutionID`, that error can overwrite a sibling's successful result.

> **Rule of thumb:** register the same custom actions on **every** object in a domain, even
> if some are no-ops for that device. It keeps domain-wide executions clean.

**3. Action names are domain-scoped on the platform.** `NewAction` publishes
`{action, domain}`. Two objects in the same domain registering the same name is fine — the
duplicate `POST` returns `ERR_ITEM_ALREADY_EXIST`, which `RegisterObject` deliberately
ignores. Two *different* drivers publishing the same `{action, domain}` pair with different
payload expectations, however, will collide on the platform. Namespace accordingly.

---

## The payload contract

### What actually arrives

The runner deserializes the incoming event into:

```go
type requestActionExecutionEventData struct {
    Payload           map[string]interface{} `json:"payload"`
    Domain            string                 `json:"domain"`
    Action            string                 `json:"action"`
    ObjectID          []string               `json:"object_id"`
    ActionExecutionID string                 `json:"id"`
}
```

then re-marshals `Payload` and hands the bytes to your handler. So:

| Platform sends | Your handler's `ctx.Payload` |
|---|---|
| `{"pattern":"wail","seconds":10}` | `{"pattern":"wail","seconds":10}` |
| `{}` | `{}` |
| no `payload` field, or `"payload": null` | **`null`** — four bytes, not `{}` |
| a JSON array or scalar | *nothing* — the whole event fails to unmarshal and is dropped with a log line; your handler is never called |

Two rules follow:

- **The payload is always a JSON object in production.** Design for that. Arrays and
  scalars are not a supported payload shape at the wire level.
- **`null` is a legal input you must handle.** `json.Unmarshal([]byte("null"), &p)`
  succeeds and leaves `p` at its zero value — no error. Required-field validation is
  therefore **your** job; you will not get an unmarshal error to lean on.

```go
func syncClock(ctx objects.CustomActionContext) (map[string]string, error) {
    var p struct {
        Timestamp string `json:"timestamp"` // RFC3339
    }
    // Note: succeeds on `null` and `{}` alike, leaving Timestamp == "".
    if err := json.Unmarshal(ctx.Payload, &p); err != nil {
        return nil, fmt.Errorf("invalid payload: %w", err)
    }
    if p.Timestamp == "" {
        p.Timestamp = time.Now().Format(time.RFC3339) // default...
        // ...or reject explicitly:
        // return nil, fmt.Errorf("timestamp is required")
    }
    // ...
    return map[string]string{"status": "ok"}, nil
}
```

### Defaults over rejection

Because nothing in the platform UI describes your payload (see
[Limitations](#limitations)), an operator triggering the action by hand will frequently
send `{}`. Prefer sensible defaults for optional fields and reject only what genuinely
cannot be guessed.

---

## Returning results and errors

The `map[string]string` you return is PUT verbatim to
`/objects/actions/executions/{ExecutionID}` as `{"result": {...}}`.

```go
return map[string]string{"status": "done", "duration_ms": "123"}, nil
```

Returning an error reports the action as failed; the runner replaces your result with a
single key:

```go
return nil, fmt.Errorf("device unreachable: %w", err)
// ⇒ {"result": {"error": "device unreachable: dial tcp ...: i/o timeout"}}
```

Guidelines:

- **Values are strings.** Format numbers and booleans yourself
  (`strconv.Itoa`, `strconv.FormatBool`, `fmt.Sprintf`).
- **Returning `(nil, nil)` is legal** and writes an empty result. Prefer at least
  `{"status": "ok"}` so an operator can tell success from a dropped execution.
- **Error strings surface to operators.** Include the device, the operation, and the
  underlying cause; skip stack traces and credentials.
- **Never return both.** On a non-nil error the result map is discarded — partial progress
  must be encoded in the error string or pushed via `ctx.Controller` before returning.

---

## Concurrency and panics

Each execution is dispatched in its own goroutine:

```go
go RunActionRoutine(obj)
```

So:

- **Handlers must be safe for concurrent invocation.** The same handler can run several
  times at once (repeated triggers, domain fan-out). Guard shared device connections or
  driver state with a mutex; do not rely on the SDK to serialize anything.
- **A panic in your handler terminates the driver process.** The goroutine has no
  `recover()` anywhere above it, so an unhandled panic is fatal for the whole driver, not
  just the action. Handlers that touch parsing, type assertions, or third-party device SDKs
  should defend themselves:

```go
func safeHandler(inner objects.CustomActionHandler) objects.CustomActionHandler {
    return func(ctx objects.CustomActionContext) (resp map[string]string, err error) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("panic in %s: %v", ctx.Action, r)
            }
        }()
        return inner(ctx)
    }
}

obj.RegisterCustomAction("switch.action.blink", safeHandler(blink))
```

- **There is no timeout and no cancellation.** The SDK will wait as long as your handler
  runs, and nothing will interrupt it. Long device calls should carry their own deadline:

```go
ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := device.CalibrateContext(ctxTimeout); err != nil { ... }
```

For anything genuinely long-running, return immediately with an acknowledgement and push
progress through the controller:

```go
obj.RegisterCustomAction("video_channel.action.reindex", func(ctx objects.CustomActionContext) (map[string]string, error) {
    go func() {
        err := reindexArchive(ctx.Object.GetMetadata().ObjectID)
        status := "completed"
        if err != nil {
            status = "failed: " + err.Error()
        }
        _ = ctx.Controller.UpdateResultAttributes(ctx.ExecutionID, map[string]string{"status": status})
    }()
    return map[string]string{"status": "started"}, nil
})
```

---

## Using the controller from a handler

`ctx.Controller` is the object's `ObjectController` — the same one passed to `Setup`. It is
assigned **inside `Setup`**, which `RegisterObject` calls *last*:

```go
func (s *switchObject) Setup(oc ObjectController) error {
    s.controller = oc
    ...
}
```

So `ctx.Controller` is **nil** before the object has been registered with the runner. In
production that is never a problem — actions only arrive after registration. In unit tests
that call `RunAction` directly it is nil, so a handler that uses the controller
unconditionally will panic in tests. Either inject a fake controller, or guard:

```go
if ctx.Controller != nil {
    _ = ctx.Controller.UpdateResultAttributes(ctx.ExecutionID, map[string]string{"phase": "connecting"})
}
```

Typical uses inside a handler:

```go
// Change the object's state as a side effect of the action.
sw := ctx.Object.(objects.SwitchObject)
_ = sw.SetState(objects.SWITCH_STATE_ON)
_ = sw.UpdateStateAttributes(map[string]string{"last_action": ctx.Action})

// Identify which object was hit during a domain-wide fan-out.
id := ctx.Object.GetMetadata().ObjectID
```

---

## Naming

Convention: `<domain-or-type>.action.<verb>`, lowercase, `snake_case` verb.

```
switch.action.blink
reader.action.sync_clock
video_channel.action.reindex
siren.action.play_pattern
```

Rules:

- **Never reuse a built-in action name** — built-ins are matched first, so your handler
  would be dead code. Check the object's `RunAction` switch, or
  [Actions Payload Reference](../doc/actions-payload-reference.md), before choosing.
- **Keep names stable.** The name is the platform-side identifier; renaming leaves the old
  action published and breaks any automation referencing it.
- **Namespace by product, not by driver instance.** Names are shared across every driver
  publishing into the same domain.

---

## Patterns

### Register the same set across a domain

```go
func registerDoorActions(d objects.DoorObject, dev *devices.DeviceManager) error {
    for name, h := range map[string]objects.CustomActionHandler{
        "door.action.pulse":     pulseHandler(dev),
        "door.action.self_test": selfTestHandler(dev),
    } {
        if err := d.RegisterCustomAction(name, h); err != nil {
            return fmt.Errorf("register %s: %w", name, err)
        }
    }
    return nil
}
```

Call it for **every** door you build, so domain-wide executions never hit a missing handler.

### Close over the device, don't look it up in the handler

```go
func pulseHandler(dev *devices.DeviceManager) objects.CustomActionHandler {
    return func(ctx objects.CustomActionContext) (map[string]string, error) {
        conn, err := dev.GetOrConnect(ctx.Object.GetMetadata().DeviceID)
        if err != nil {
            return nil, fmt.Errorf("connect: %w", err)
        }
        // ...
        return map[string]string{"status": "pulsed"}, nil
    }
}
```

### Validate once, in a helper

```go
func decode[T any](payload []byte) (T, error) {
    var v T
    if len(payload) == 0 || string(payload) == "null" {
        return v, nil // caller applies defaults
    }
    err := json.Unmarshal(payload, &v)
    return v, err
}
```

---

## Full example

```go
func NewSiren(objectID, deviceID string, dev *devices.DeviceManager) (objects.SwitchObject, error) {
    sw := objects.NewSwitchObject(objects.NewSwitchObjectParams{
        Metadata: objects.ObjectMetadata{
            ObjectID: objectID,
            Name:     "Siren",
            Domain:   "siren",
            DeviceID: deviceID,
        },
        TurnOnMethod:  func(o objects.RegistrableObject, oc objects.ObjectController) error { /* ... */ return nil },
        TurnOffMethod: func(o objects.RegistrableObject, oc objects.ObjectController) error { /* ... */ return nil },
    })

    // Custom action: play a specific tone pattern for N seconds.
    err := sw.RegisterCustomAction("siren.action.play_pattern", func(ctx objects.CustomActionContext) (map[string]string, error) {
        var p struct {
            Pattern string `json:"pattern"`
            Seconds int    `json:"seconds"`
        }
        // Succeeds on `null` and `{}` too — defaults below cover those.
        if err := json.Unmarshal(ctx.Payload, &p); err != nil {
            return nil, fmt.Errorf("invalid play_pattern payload: %w", err)
        }
        if p.Pattern == "" {
            p.Pattern = "wail"
        }
        if p.Seconds <= 0 {
            p.Seconds = 5
        }

        conn, err := dev.GetOrConnect(deviceID)
        if err != nil {
            return nil, fmt.Errorf("siren %s unreachable: %w", deviceID, err)
        }
        if err := conn.PlayPattern(p.Pattern, p.Seconds); err != nil {
            return nil, fmt.Errorf("play_pattern failed: %w", err)
        }

        _ = sw.UpdateStateAttributes(map[string]string{"last_pattern": p.Pattern})

        return map[string]string{
            "status":  "playing",
            "pattern": p.Pattern,
            "seconds": strconv.Itoa(p.Seconds),
        }, nil
    })
    if err != nil {
        return nil, err
    }

    return sw, nil
}

// ...at startup:
siren, err := NewSiren("siren-1", "dev-1", dev)
if err != nil {
    log.Fatal(err)
}
if err := client.RegisterObject(siren); err != nil { // actions are advertised here
    log.Fatal(err)
}
```

The platform triggers it with an event shaped like:

```json
{
  "id": "exec-8f2c",
  "domain": "siren",
  "action": "siren.action.play_pattern",
  "object_id": ["siren-1"],
  "payload": { "pattern": "wail", "seconds": 10 }
}
```

and the driver reports back:

```json
PUT /objects/actions/executions/exec-8f2c
{ "result": { "status": "playing", "pattern": "wail", "seconds": "10" } }
```

---

## Testing

`RunAction` is directly callable, so custom actions unit-test without any platform:

```go
func TestPlayPattern(t *testing.T) {
    sw, err := NewSiren("s1", "d1", fakeDeviceManager())
    require.NoError(t, err)

    resp, err := sw.RunAction("exec-1", "siren.action.play_pattern", []byte(`{"pattern":"wail","seconds":3}`))
    require.NoError(t, err)
    assert.Equal(t, "playing", resp["status"])
    assert.Equal(t, "3", resp["seconds"])
}

// Cover the shapes the platform really sends.
func TestPlayPattern_Defaults(t *testing.T) {
    sw, _ := NewSiren("s1", "d1", fakeDeviceManager())

    for _, payload := range []string{`{}`, `null`, ``} {
        resp, err := sw.RunAction("exec-1", "siren.action.play_pattern", []byte(payload))
        require.NoError(t, err, "payload %q", payload)
        assert.Equal(t, "wail", resp["pattern"])
    }
}

// The action must be advertised, or the platform can never trigger it.
func TestPlayPattern_IsAdvertised(t *testing.T) {
    sw, _ := NewSiren("s1", "d1", fakeDeviceManager())

    var found bool
    for _, a := range sw.GetAvailableActions() {
        if a.Action == "siren.action.play_pattern" {
            found = true
            assert.Equal(t, "siren", a.Domain)
        }
    }
    assert.True(t, found)
}
```

Remember `ctx.Controller` is nil in these tests — `Setup` has not run. Inject a fake
controller via `sw.Setup(fakeController)` if the handler needs one.

The SDK itself ships regression and adversarial tests for the feature
([`pkg/objects/custom_actions_test.go`](../pkg/objects/custom_actions_test.go),
[`pkg/objects/custom_actions_robustness_test.go`](../pkg/objects/custom_actions_robustness_test.go)):

- built-in action sets stay unchanged when custom actions are added;
- a custom action is dispatched for nil / empty / malformed / non-object payloads across
  every object type, without panicking, receiving the exact bytes sent;
- registration validation (empty name, nil handler, duplicates);
- concurrent registration and dispatch under `-race`.

---

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| Action never appears in the platform UI | Registered *after* `RegisterObject`; `GetAvailableActions()` is read only during registration. |
| Handler never runs; result shows `action <name> not found` | Name collides with a built-in (built-ins match first), or the execution fanned out to a sibling object in the same domain that lacks the handler. |
| Handler never runs; nothing in the result at all | Domain mismatch — the runner logs `no objects found for domain`. Check `ObjectMetadata.Domain` against the event's `domain`. |
| Fields arrive empty even though the operator filled them in | Payload arrived as `null`/`{}` — `json.Unmarshal` returned no error. Validate required fields explicitly. |
| Driver process dies when the action runs | Panic in the handler. The dispatch goroutine has no `recover()`; wrap the handler. |
| Result shows a sibling's error, not your success | Domain-wide execution: several objects share one `ExecutionID` and the last `UpdateResultAttributes` wins. |
| `RegisterCustomAction` returns "already registered" | The same name registered twice on one object — often a constructor called twice on a shared object. |

Driver logs (zap) are the fastest confirmation: the runner logs `running action` with the
action, domain, object IDs, and payload before dispatching, then `action executed` or
`action execution error`.

---

## Limitations

Known and by design as of `v0.7.78`:

- **No payload schema is published.** `ObjectAction` carries only `{action, domain}`. The
  platform cannot render a typed form, validate input, or document parameters. Document the
  expected payload in your driver's integration manual and default aggressively.
- **No display metadata.** No label, description, icon, or grouping — the raw action name
  is what operators see. Choose readable names.
- **No deregistration.** Handlers cannot be removed once registered; build the object with
  the exact set it should expose.
- **No per-execution timeout or cancellation.** Enforce deadlines yourself.
- **No automatic panic isolation.** Wrap handlers that can panic.
- **Results are `map[string]string` only.** Nested structures must be serialized into a
  string value.

---

## FAQ

**Can I remove a custom action after registering it?**
No. Build the object with the actions it should expose and register them once before
`RegisterObject`.

**Can two objects share the same custom action name?**
Yes, and within a single domain you generally *should* — see
[Routing](#routing-actions-are-dispatched-by-domain). Duplicate `NewAction` calls return
`ERR_ITEM_ALREADY_EXIST`, which `RegisterObject` ignores on purpose.

**Can I register a custom action after `RegisterObject`?**
It will dispatch if the platform somehow knows the name, but it is never advertised, so
nothing in the UI can trigger it. Treat registration as build-time only.

**Does this change how built-in actions behave?**
No. Built-in actions are matched first and are completely unaffected.

**Can a custom action return binary data or a file?**
Not through the result map. Upload it through the event/media path and return a reference
(URL, ID) as a string.

**How do I tell which object a fan-out execution hit?**
`ctx.Object.GetMetadata().ObjectID`.

**What SDK version do I need?**
`v0.7.78` or newer.

---

See also: [Understanding Objects](objects.md) ·
[`doc/quick-start/03-understanding-objects.md`](../doc/quick-start/03-understanding-objects.md) ·
[Actions Payload Reference](../doc/actions-payload-reference.md)
