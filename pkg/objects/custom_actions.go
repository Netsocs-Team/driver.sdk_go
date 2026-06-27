package objects

import (
	"fmt"
	"sync"
)

// CustomActionContext is passed to a custom action handler when DriversHub
// triggers a driver-defined action. It carries everything the handler needs to
// process the request and report results back to the platform.
type CustomActionContext struct {
	ExecutionID string            // action execution id from the platform
	Action      string            // the custom action name being invoked
	Payload     []byte            // raw JSON payload sent by DriversHub
	Object      RegistrableObject // the object the action was invoked on
	Controller  ObjectController  // SDK controller for state / result updates
}

// CustomActionHandler handles a single custom action invocation. The returned
// map is sent back to the platform as the action execution result.
type CustomActionHandler func(ctx CustomActionContext) (map[string]string, error)

// CustomActionRegistrar is implemented by every built-in object so drivers can
// attach their own actions without modifying the SDK. Register actions before
// calling ObjectRunner.RegisterObject so they are advertised to the platform.
type CustomActionRegistrar interface {
	RegisterCustomAction(action string, handler CustomActionHandler) error
}

// customActions is an embeddable registry of driver-defined custom actions.
// Built-in object types embed it (anonymously) to gain RegisterCustomAction and
// the dispatch/advertise helpers. The zero value is ready to use; the map is
// initialized lazily. Objects are always used through pointers, so embedding the
// mutex by value is safe.
type customActions struct {
	mu       sync.RWMutex
	handlers map[string]CustomActionHandler
}

// RegisterCustomAction registers a driver-defined action handler under the given
// name. It must be called before the object is registered with the runner so the
// action is published to the platform. It returns an error if the name is empty,
// the handler is nil, or the action is already registered.
func (c *customActions) RegisterCustomAction(action string, handler CustomActionHandler) error {
	if action == "" {
		return fmt.Errorf("custom action name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("custom action handler cannot be nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.handlers == nil {
		c.handlers = make(map[string]CustomActionHandler)
	}
	if _, exists := c.handlers[action]; exists {
		return fmt.Errorf("custom action %q already registered", action)
	}
	c.handlers[action] = handler
	return nil
}

// customActionList returns the registered custom actions as ObjectActions for the
// given domain, so they are advertised to the platform alongside the built-in
// actions through GetAvailableActions.
func (c *customActions) customActionList(domain string) []ObjectAction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	list := make([]ObjectAction, 0, len(c.handlers))
	for action := range c.handlers {
		list = append(list, ObjectAction{Action: action, Domain: domain})
	}
	return list
}

// runCustomAction dispatches to a registered custom handler. The boolean reports
// whether a handler was found, so callers can fall back to their usual
// "action not found" handling.
func (c *customActions) runCustomAction(ctx CustomActionContext) (map[string]string, bool, error) {
	c.mu.RLock()
	handler, ok := c.handlers[ctx.Action]
	c.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	resp, err := handler(ctx)
	return resp, true, err
}

// dispatchCustom is the standard tail for an object's RunAction default case: it
// tries a registered custom handler and otherwise returns the canonical
// "action not found" error.
func (c *customActions) dispatchCustom(this RegistrableObject, oc ObjectController, id, action string, payload []byte) (map[string]string, error) {
	if resp, ok, err := c.runCustomAction(CustomActionContext{
		ExecutionID: id,
		Action:      action,
		Payload:     payload,
		Object:      this,
		Controller:  oc,
	}); ok {
		return resp, err
	}
	return nil, fmt.Errorf("action %s not found", action)
}
