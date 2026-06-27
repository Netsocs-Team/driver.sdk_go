# Custom Actions

Custom actions let a **driver** add its own actions to any built-in object, on top of
the predefined ones the SDK ships with — without forking the SDK or implementing a new
object type.

Use them when a device supports an operation that the built-in action set doesn't cover:
reboot, calibrate, blink an LED, run a self-test, sync a clock, trigger a relay pulse, etc.

- **Available since:** SDK `v0.7.78`
- **Source:** [`pkg/objects/custom_actions.go`](../pkg/objects/custom_actions.go)
- **Works on:** every built-in object type (`switch`, `door`, `lock`, `sensor`, `speaker`,
  `microphone`, `video_channel`, `video_engine`, `alarm_panel`, `reader`, `octopus`,
  `notifier`, `gps_tracker`, `person`, `relative_zone`, `relative_tracker`).

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
        return nil, err
    }
    // ... drive the device ...
    return map[string]string{"status": "blinked"}, nil
})
if err != nil {
    log.Fatal(err)
}

// 2. Now register the object — the custom action is published to the platform here.
runner.RegisterObject(sw)
```

---

## How it works

Custom actions ride on the two contracts every object already exposes, so nothing in the
dispatch pipeline changes:

1. **Advertise** — `GetAvailableActions()` returns the built-in actions **plus** every
   registered custom action. The runner reads this once during `RegisterObject` and
   publishes the actions to the platform (`CreateObject` + `NewAction`).
2. **Dispatch** — when the platform triggers an action, `object_runner` calls
   `RunAction(...)`. Built-in actions match the object's `switch` first; anything that
   falls through to the `default` case is looked up in the custom-action registry.

```
DriversHub ──REQUEST_ACTION_EXECUTION──► object_runner
                                              │
                                              ▼
                                       obj.RunAction(id, action, payload)
                                              │
                            ┌─────────────────┴───────────────────┐
                            │ switch action {                     │
                            │   case <built-in>: ...  ◄── built-in actions win
                            │   default:                          │
                            │     dispatchCustom(...) ◄── your custom handler
                            │ }                                   │
                            └─────────────────────────────────────┘
```

Because built-ins are matched first, **a custom action can never shadow a built-in one**.

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

// Everything the handler needs.
type CustomActionContext struct {
    ExecutionID string            // action execution id from the platform
    Action      string            // the custom action name being invoked
    Payload     []byte            // raw JSON payload sent by DriversHub
    Object      RegistrableObject // the object the action was invoked on
    Controller  ObjectController  // for SetState / UpdateResultAttributes / etc.
}
```

### `RegisterCustomAction(action string, handler CustomActionHandler) error`

Registers a handler under `action`. Returns an error when:

| Condition | Error |
|---|---|
| `action == ""` | `custom action name cannot be empty` |
| `handler == nil` | `custom action handler cannot be nil` |
| `action` already registered | `custom action "<name>" already registered` |

The registry is safe for concurrent use.

---

## Step by step

### 1. Pick an action name

A namespaced, lowercase string is the convention, e.g. `switch.action.blink`,
`reader.action.sync_clock`. **Do not reuse a built-in action name** — built-ins take
precedence, so a colliding name would never reach your handler.

### 2. Register before `RegisterObject`

The runner reads `GetAvailableActions()` **once**, during `RegisterObject`, to publish the
action to the platform. Register all custom actions before that call:

```go
obj := objects.NewReaderObject(params)
obj.RegisterCustomAction("reader.action.sync_clock", syncClock)
obj.RegisterCustomAction("reader.action.self_test", selfTest)

runner.RegisterObject(obj) // <-- actions are advertised here
```

Registering after this point still dispatches correctly if the platform somehow knows the
name, but the action will **not** be advertised in the UI.

### 3. Parse the payload yourself

The payload is opaque to the SDK: it forwards the raw JSON bytes as-is. Your handler owns
parsing. Always send a JSON **object** as the payload.

```go
func syncClock(ctx objects.CustomActionContext) (map[string]string, error) {
    var p struct {
        Timestamp string `json:"timestamp"` // RFC3339
    }
    if err := json.Unmarshal(ctx.Payload, &p); err != nil {
        return nil, fmt.Errorf("invalid payload: %w", err)
    }
    // ...
    return map[string]string{"status": "ok"}, nil
}
```

### 4. Return a result (or an error)

The returned `map[string]string` is sent back to the platform as the action execution
result. Returning an error reports the action as failed (the error string is stored under
`error`).

```go
return map[string]string{"status": "done", "duration_ms": "123"}, nil
// or
return nil, fmt.Errorf("device unreachable")
```

You can also drive the object from inside the handler via `ctx.Object` / `ctx.Controller`:

```go
sw := ctx.Object.(objects.SwitchObject)
_ = sw.UpdateStateAttributes(map[string]string{"last_action": ctx.Action})
```

---

## Full example

```go
func NewSiren(objectID, deviceID string, dev *devices.DeviceManager) objects.SwitchObject {
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
    sw.RegisterCustomAction("siren.action.play_pattern", func(ctx objects.CustomActionContext) (map[string]string, error) {
        var p struct {
            Pattern  string `json:"pattern"`
            Seconds  int    `json:"seconds"`
        }
        if err := json.Unmarshal(ctx.Payload, &p); err != nil {
            return nil, fmt.Errorf("invalid play_pattern payload: %w", err)
        }
        if p.Seconds <= 0 {
            p.Seconds = 5
        }

        // device, err := dev.GetOrConnect(...) ; device.PlayPattern(p.Pattern, p.Seconds)

        return map[string]string{
            "status":  "playing",
            "pattern": p.Pattern,
            "seconds": fmt.Sprintf("%d", p.Seconds),
        }, nil
    })

    return sw
}
```

The platform triggers it with a payload like:

```json
{ "pattern": "wail", "seconds": 10 }
```

---

## Rules & gotchas

- **Register before `RegisterObject`.** Otherwise the action isn't advertised to the platform.
- **Built-in actions win.** Don't reuse a built-in action name; the built-in handler is
  dispatched first and your custom handler would never run.
- **Payload is opaque and should be a JSON object.** The SDK does not validate or unmarshal
  it — your handler does. Sending a non-object (array, bare string/number) is discouraged;
  the handler still receives the raw bytes, but most drivers expect an object.
- **Each name is unique per object.** Re-registering the same name returns an error.
- **A handler panic propagates** like any other action handler — guard your own code.

---

## Testing

The SDK ships regression and adversarial tests for this feature
([`pkg/objects/custom_actions_test.go`](../pkg/objects/custom_actions_test.go),
[`pkg/objects/custom_actions_robustness_test.go`](../pkg/objects/custom_actions_robustness_test.go)):

- built-in action sets stay unchanged when custom actions are added;
- a custom action is dispatched for nil / empty / malformed / non-object payloads across
  every object type, without panicking, receiving the exact bytes sent;
- registration validation (empty name, nil handler, duplicates);
- concurrent registration and dispatch under `-race`.

A minimal test for your own custom action:

```go
func TestBlink(t *testing.T) {
    sw := NewSiren("s1", "d1", nil)

    resp, err := sw.RunAction("exec-1", "siren.action.play_pattern", []byte(`{"pattern":"wail","seconds":3}`))
    require.NoError(t, err)
    assert.Equal(t, "playing", resp["status"])
}
```

---

## FAQ

**Can I remove a custom action after registering it?**
No. Build the object with the actions it should expose and register them once before
`RegisterObject`.

**Can two objects share the same custom action name?**
Yes. Names are scoped per object instance (the `ObjectAction.Domain` is the object's domain).

**Does this change how built-in actions behave?**
No. Built-in actions are matched first and are completely unaffected.

**What Go version / SDK version do I need?**
SDK `v0.7.78` or newer.

---

See also: [Understanding Objects](objects.md) ·
[`doc/quick-start/03-understanding-objects.md`](../doc/quick-start/03-understanding-objects.md) ·
[Actions Payload Reference](../doc/actions-payload-reference.md)
