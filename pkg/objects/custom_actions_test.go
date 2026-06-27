package objects

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterCustomAction_validation(t *testing.T) {
	var reg customActions

	assert.Error(t, reg.RegisterCustomAction("", func(CustomActionContext) (map[string]string, error) { return nil, nil }),
		"empty action name should error")
	assert.Error(t, reg.RegisterCustomAction("x.action.y", nil),
		"nil handler should error")

	require.NoError(t, reg.RegisterCustomAction("x.action.y", func(CustomActionContext) (map[string]string, error) { return nil, nil }))
	assert.Error(t, reg.RegisterCustomAction("x.action.y", func(CustomActionContext) (map[string]string, error) { return nil, nil }),
		"duplicate action name should error")
}

func TestCustomActionList(t *testing.T) {
	var reg customActions
	require.NoError(t, reg.RegisterCustomAction("a", func(CustomActionContext) (map[string]string, error) { return nil, nil }))
	require.NoError(t, reg.RegisterCustomAction("b", func(CustomActionContext) (map[string]string, error) { return nil, nil }))

	list := reg.customActionList("mydomain")
	require.Len(t, list, 2)

	names := map[string]bool{}
	for _, a := range list {
		assert.Equal(t, "mydomain", a.Domain)
		names[a.Action] = true
	}
	assert.True(t, names["a"])
	assert.True(t, names["b"])
}

func newTestSwitch() SwitchObject {
	return NewSwitchObject(NewSwitchObjectParams{
		Metadata:     ObjectMetadata{ObjectID: "sw-1", Domain: "switch"},
		TurnOnMethod: func(RegistrableObject, ObjectController) error { return nil },
	})
}

func TestCustomAction_advertisedInAvailableActions(t *testing.T) {
	sw := newTestSwitch()
	require.NoError(t, sw.RegisterCustomAction("switch.action.blink", func(CustomActionContext) (map[string]string, error) {
		return nil, nil
	}))

	actions := sw.GetAvailableActions()
	have := map[string]bool{}
	for _, a := range actions {
		assert.Equal(t, "switch", a.Domain)
		have[a.Action] = true
	}
	assert.True(t, have[SWITCH_ACTION_TURN_ON], "built-in turn_on should still be advertised")
	assert.True(t, have[SWITCH_ACTION_TURN_OFF], "built-in turn_off should still be advertised")
	assert.True(t, have["switch.action.blink"], "custom action should be advertised")
}

func TestCustomAction_dispatch(t *testing.T) {
	sw := newTestSwitch()

	var gotCtx CustomActionContext
	require.NoError(t, sw.RegisterCustomAction("switch.action.blink", func(ctx CustomActionContext) (map[string]string, error) {
		gotCtx = ctx
		return map[string]string{"blinked": "true"}, nil
	}))

	payload := []byte(`{"times":3}`)
	resp, err := sw.RunAction("exec-1", "switch.action.blink", payload)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"blinked": "true"}, resp)

	// the handler received the untouched payload and context
	assert.Equal(t, "exec-1", gotCtx.ExecutionID)
	assert.Equal(t, "switch.action.blink", gotCtx.Action)
	assert.Equal(t, payload, gotCtx.Payload)
	assert.NotNil(t, gotCtx.Object)
}

func TestCustomAction_handlerErrorPropagates(t *testing.T) {
	sw := newTestSwitch()
	require.NoError(t, sw.RegisterCustomAction("switch.action.fail", func(CustomActionContext) (map[string]string, error) {
		return nil, fmt.Errorf("boom")
	}))

	_, err := sw.RunAction("exec", "switch.action.fail", []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestCustomAction_unknownActionStillNotFound(t *testing.T) {
	sw := newTestSwitch()
	_, err := sw.RunAction("exec", "switch.action.does_not_exist", []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCustomAction_builtinTakesPrecedence(t *testing.T) {
	builtinCalled := false
	customCalled := false

	sw := NewSwitchObject(NewSwitchObjectParams{
		Metadata:     ObjectMetadata{ObjectID: "sw-1", Domain: "switch"},
		TurnOnMethod: func(RegistrableObject, ObjectController) error { builtinCalled = true; return nil },
	})

	// register a custom action that shadows a built-in name
	require.NoError(t, sw.RegisterCustomAction(SWITCH_ACTION_TURN_ON, func(CustomActionContext) (map[string]string, error) {
		customCalled = true
		return nil, nil
	}))

	_, err := sw.RunAction("exec", SWITCH_ACTION_TURN_ON, []byte(`{}`))
	require.NoError(t, err)
	assert.True(t, builtinCalled, "built-in handler should run")
	assert.False(t, customCalled, "custom handler must not shadow a built-in action")
}
