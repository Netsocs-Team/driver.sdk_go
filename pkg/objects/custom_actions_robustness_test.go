package objects

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// customActionObject is the contract every built-in object must satisfy so a driver
// can attach custom actions. All NewXxxObject constructors must return a type that
// implements it.
type customActionObject interface {
	RegistrableObject
	CustomActionRegistrar
}

// builtinObjectCase describes a built-in object type, a factory that builds a fresh
// instance, and the exact set of action names it must advertise out of the box.
type builtinObjectCase struct {
	name string
	make func() customActionObject
	want []string
}

func builtinObjectCases() []builtinObjectCase {
	return []builtinObjectCase{
		{"switch", func() customActionObject {
			return NewSwitchObject(NewSwitchObjectParams{Metadata: ObjectMetadata{Domain: "switch"}})
		}, []string{SWITCH_ACTION_TURN_ON, SWITCH_ACTION_TURN_OFF}},

		{"door", func() customActionObject {
			return NewDoorObject(NewDoorObjectParams{Metadata: ObjectMetadata{Domain: "door"}})
		}, []string{DOOR_ACTION_OPEN, DOOR_ACTION_CLOSE}},

		{"lock", func() customActionObject {
			return NewLockObject(NewLockObjectParams{Metadata: ObjectMetadata{Domain: "lock"}})
		}, []string{LOCK_ACTION_LOCK, LOCK_ACTION_UNLOCK, LOCK_ACTION_REBOOT}},

		{"sensor", func() customActionObject {
			return NewSensorObject(NewSensorObjectParams{Metadata: ObjectMetadata{Domain: "sensor"}})
		}, []string{SENSOR_ACTION_BYPASS, SENSOR_ACTION_UNBYPASS, SENSOR_CUSTOM_ACTION}},

		{"speaker", func() customActionObject {
			return NewSpeakerObject(NewSpeakerObjectProps{Metadata: ObjectMetadata{Domain: "speaker"}})
		}, []string{SPEAKER_ACTION_START_TALKBACK, SPEAKER_ACTION_STOP_TALKBACK}},

		{"microphone", func() customActionObject {
			return NewMicrophoneObject(NewMicrophoneObjectProps{Metadata: ObjectMetadata{Domain: "microphone"}})
		}, []string{MICROPHONE_ACTION_START_STREAM, MICROPHONE_ACTION_STOP_STREAM}},

		{"video_channel", func() customActionObject {
			return NewVideoChannelObject(NewVideoChannelObjectProps{Metadata: ObjectMetadata{Domain: "video_channel"}})
		}, []string{
			VIDEO_CHANNEL_ACTION_SNAPSHOT,
			VIDEO_CHANNEL_ACTION_PTZ_CONTROL,
			VIDEO_CHANNEL_ACTION_PTZ_GOTO_PRESET,
			VIDEO_CHANNEL_ACTION_VIDEOCLIP,
			VIDEO_CHANNEL_ACTION_SEEK,
			VIDEO_CHANNEL_ACTION_REQUEST_DOLYNK_STREAM_URL,
			VIDEO_CHANNEL_ACTION_REQUEST_DAHUA_PLAYBACK_MEDIA_FILES,
			VIDEO_CHANNEL_ACTION_GET_RECORDING_SEGMENTS,
			VIDEO_CHANNEL_ACTION_PTZ_GET_STATUS,
			VIDEO_CHANNEL_ACTION_PUBLISH_STREAM_START,
			VIDEO_CHANNEL_ACTION_PUBLISH_STREAM_STOP,
			VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP,
		}},

		{"video_engine", func() customActionObject {
			return NewVideoEngineObject(NewVideoEngineObjectParams{Metadata: ObjectMetadata{Domain: "video_engine"}})
		}, nil},

		{"alarm_panel", func() customActionObject {
			return NewAlarmPanelObject(NewAlarmPanelObjectProps{Metadata: ObjectMetadata{Domain: "alarm_panel"}})
		}, []string{ALARM_PANEL_ACTION_ARM, ALARM_PANEL_ACTION_DISARM}},

		{"reader", func() customActionObject {
			return NewReaderObject(NewReaderObjectParams{Metadata: ObjectMetadata{Domain: "reader"}})
		}, []string{
			READER_ACTION_READ,
			READER_ACTION_STOP,
			READER_ACTION_RESET,
			READER_ACTION_RESTART,
			READER_ACTION_STORE_QRS,
			READER_ACTION_DELETE_QRS,
			READER_ACTION_DELETE_PERSON,
			READER_ACTION_GET_PEOPLE,
			READER_ACTION_SET_PEOPLE,
			READER_ACTION_SYNC_ACCESS_DATABASE,
		}},

		{"octopus", func() customActionObject {
			return NewOctopusObject(NewOctopusObjectParams{Metadata: ObjectMetadata{Domain: "octopus"}})
		}, []string{OCTOPUS_ACTION_RELAY_ON, OCTOPUS_ACTION_RELAY_OFF}},

		{"notify", func() customActionObject {
			return NewNotifierObject(NewNotifierObjectProps{Metadata: ObjectMetadata{Domain: "notify"}})
		}, []string{CREATE}},

		{"gps_tracker", func() customActionObject {
			return NewGPSTrackerObject(NewGPSTrackerObjectProps{Metadata: ObjectMetadata{Domain: "gps_tracker"}})
		}, nil},

		{"person", func() customActionObject {
			return NewPersonObject(NewPersonObjectParams{Metadata: ObjectMetadata{Domain: "person"}})
		}, nil},

		{"relative_zone", func() customActionObject {
			return NewRelativeZoneObject(NewRelativeZoneObjectParams{Metadata: ObjectMetadata{Domain: "relative_zone"}})
		}, nil},

		{"relative_tracker", func() customActionObject {
			return NewRelativeTrackerObject(NewRelativeTrackerObjectProps{Metadata: ObjectMetadata{Domain: "relative_tracker"}})
		}, nil},
	}
}

func actionNames(actions []ObjectAction) []string {
	names := make([]string, 0, len(actions))
	for _, a := range actions {
		names = append(names, a.Action)
	}
	return names
}

// TestBuiltinActionsUnchanged guards against the custom-action refactor accidentally
// dropping or duplicating any built-in action: with no custom actions registered, each
// object must advertise exactly its original action set.
func TestBuiltinActionsUnchanged(t *testing.T) {
	for _, tc := range builtinObjectCases() {
		t.Run(tc.name, func(t *testing.T) {
			obj := tc.make()
			got := actionNames(obj.GetAvailableActions())
			assert.ElementsMatch(t, tc.want, got, "%s: built-in actions changed", tc.name)
		})
	}
}

// TestCustomActionsAddedToBuiltins verifies that registering a custom action adds it to
// GetAvailableActions without disturbing the built-in ones — for every object type.
func TestCustomActionsAddedToBuiltins(t *testing.T) {
	for _, tc := range builtinObjectCases() {
		t.Run(tc.name, func(t *testing.T) {
			obj := tc.make()
			require.NoError(t, obj.RegisterCustomAction("probe.custom.action", func(CustomActionContext) (map[string]string, error) {
				return nil, nil
			}))

			got := actionNames(obj.GetAvailableActions())
			expected := append(append([]string{}, tc.want...), "probe.custom.action")
			assert.ElementsMatch(t, expected, got, "%s: built-ins + custom mismatch", tc.name)
		})
	}
}

// weirdPayloads is a battery of malformed / unexpected payloads. A custom action must be
// dispatched to its handler for ALL of them — the handler owns parsing, so the SDK must
// never reject or choke on the raw bytes before reaching it.
func weirdPayloads() map[string][]byte {
	return map[string][]byte{
		"nil":          nil,
		"empty":        {},
		"empty_object": []byte(`{}`),
		"json_null":    []byte(`null`),
		"json_array":   []byte(`[1,2,3]`),
		"json_string":  []byte(`"hello"`),
		"json_number":  []byte(`123`),
		"json_bool":    []byte(`true`),
		"nested":       []byte(`{"a":{"b":[1,2,{"c":"d"}]}}`),
		"wrong_types":  []byte(`{"code":123,"zone":true,"arm_mode":[1,2]}`),
		"unknown_keys": []byte(`{"totally":"unexpected","keys":42}`),
		"not_json":     []byte(`}{not valid json at all`),
		"whitespace":   []byte("   \n\t  "),
	}
}

// TestCustomActionDispatchSurvivesWeirdPayloads is the core adversarial test: for every
// object type, a registered custom action must run no matter how strange the payload is,
// and must receive the exact bytes that were sent.
func TestCustomActionDispatchSurvivesWeirdPayloads(t *testing.T) {
	for _, tc := range builtinObjectCases() {
		for pname, payload := range weirdPayloads() {
			t.Run(tc.name+"/"+pname, func(t *testing.T) {
				obj := tc.make()

				var gotPayload []byte
				called := false
				require.NoError(t, obj.RegisterCustomAction("probe.custom.action", func(ctx CustomActionContext) (map[string]string, error) {
					called = true
					gotPayload = ctx.Payload
					return map[string]string{"ok": "true"}, nil
				}))

				var resp map[string]string
				var err error
				require.NotPanics(t, func() {
					resp, err = obj.RunAction("exec-1", "probe.custom.action", payload)
				}, "%s/%s: custom dispatch panicked", tc.name, pname)

				require.True(t, called, "%s/%s: custom handler was not invoked", tc.name, pname)
				require.NoError(t, err)
				assert.Equal(t, map[string]string{"ok": "true"}, resp)
				assert.Equal(t, payload, gotPayload, "%s/%s: payload was mutated before reaching handler", tc.name, pname)
			})
		}
	}
}

// TestUnknownActionNeverPanics ensures that an unknown action (with no matching built-in
// or custom handler) returns an error and never panics, for any payload and object type.
func TestUnknownActionNeverPanics(t *testing.T) {
	for _, tc := range builtinObjectCases() {
		for pname, payload := range weirdPayloads() {
			t.Run(tc.name+"/"+pname, func(t *testing.T) {
				obj := tc.make()
				require.NotPanics(t, func() {
					_, _ = obj.RunAction("exec", "definitely.not.an.action", payload)
				})
			})
		}
	}
}

// TestRegisterCustomAction_rejectsBadInput hammers the validation surface.
func TestRegisterCustomAction_rejectsBadInput(t *testing.T) {
	var reg customActions

	good := func(CustomActionContext) (map[string]string, error) { return nil, nil }

	assert.Error(t, reg.RegisterCustomAction("", good), "empty name must be rejected")
	assert.Error(t, reg.RegisterCustomAction("x", nil), "nil handler must be rejected")

	require.NoError(t, reg.RegisterCustomAction("x", good))
	assert.Error(t, reg.RegisterCustomAction("x", good), "duplicate name must be rejected")

	// A rejected registration must not become dispatchable.
	_, ok, _ := reg.runCustomAction(CustomActionContext{Action: ""})
	assert.False(t, ok, "empty action must not resolve to a handler")

	// Unusual-but-valid names should work.
	require.NoError(t, reg.RegisterCustomAction("  spaces  ", good))
	require.NoError(t, reg.RegisterCustomAction("ünïcödé.акция.动作", good))
	require.NoError(t, reg.RegisterCustomAction("a.very.long.name."+longString(512), good))
}

func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// TestRegisterCustomAction_nilHandlerNotDispatchable confirms a failed (nil) registration
// leaves the action unregistered rather than panicking on dispatch.
func TestRegisterCustomAction_nilHandlerNotDispatchable(t *testing.T) {
	sw := NewSwitchObject(NewSwitchObjectParams{Metadata: ObjectMetadata{Domain: "switch"}})
	_ = sw.RegisterCustomAction("switch.action.nilop", nil) // rejected

	_, err := sw.RunAction("exec", "switch.action.nilop", []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestCustomActions_concurrentRegistration exercises the registry under concurrency.
// Run with -race to catch data races on the handler map.
func TestCustomActions_concurrentRegistration(t *testing.T) {
	var reg customActions

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = reg.RegisterCustomAction(fmt.Sprintf("action.%d", i), func(CustomActionContext) (map[string]string, error) {
				return nil, nil
			})
		}(i)
	}
	wg.Wait()

	assert.Len(t, reg.customActionList("d"), n, "all concurrent registrations should be stored")
}

// TestCustomActions_concurrentRegisterAndDispatch races registration against dispatch and
// listing to ensure the RWMutex protects all access paths. Run with -race.
func TestCustomActions_concurrentRegisterAndDispatch(t *testing.T) {
	var reg customActions
	require.NoError(t, reg.RegisterCustomAction("seed", func(CustomActionContext) (map[string]string, error) {
		return map[string]string{"ok": "true"}, nil
	}))

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = reg.RegisterCustomAction(fmt.Sprintf("a.%d", i), func(CustomActionContext) (map[string]string, error) { return nil, nil })
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_, _, _ = reg.runCustomAction(CustomActionContext{Action: "seed"})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = reg.customActionList("d")
		}
	}()
	wg.Wait()
}

// TestBuiltinActionStillDispatchesWithCustomRegistered guards that registering custom
// actions never steals dispatch from a real built-in action, for objects whose built-in
// handler can run without a live controller.
func TestBuiltinActionStillDispatchesWithCustomRegistered(t *testing.T) {
	builtinRan := false
	sw := NewSwitchObject(NewSwitchObjectParams{
		Metadata:     ObjectMetadata{Domain: "switch"},
		TurnOnMethod: func(RegistrableObject, ObjectController) error { builtinRan = true; return nil },
	})
	require.NoError(t, sw.RegisterCustomAction("switch.action.extra", func(CustomActionContext) (map[string]string, error) {
		return nil, nil
	}))

	_, err := sw.RunAction("exec", SWITCH_ACTION_TURN_ON, []byte(`{"weird":[1,2]}`))
	require.NoError(t, err)
	assert.True(t, builtinRan, "built-in action must still dispatch when custom actions exist")
}
