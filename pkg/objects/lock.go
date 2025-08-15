package objects

import "errors"

const LOCK_STATE_JAMMED = "jammed"
const LOCK_STATE_OPEN = "open"
const LOCK_STATE_OPENING = "opening"
const LOCK_STATE_LOCKED = "locked"
const LOCK_STATE_LOCKING = "locking"
const LOCK_STATE_UNLOCKED = "unlocked"
const LOCK_STATE_UNLOCKING = "unlocking"
const LOCK_STATE_UNKNOWN = "unknown"

const LOCK_ACTION_LOCK = "lock"
const LOCK_ACTION_UNLOCK = "unlock"

type LockObject interface {
	RegistrableObject
}

type lockObject struct {
	metadata     ObjectMetadata
	lockMethod   func(this LockObject, controller ObjectController) (map[string]string, error)
	unlockMethod func(this LockObject, controller ObjectController) (map[string]string, error)
	setup        func(this LockObject, controller ObjectController) error
	controller   ObjectController
}

// GetAvailableActions implements LockObject.
func (d *lockObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: LOCK_ACTION_LOCK,
			Domain: d.metadata.Domain,
		},
		{
			Action: LOCK_ACTION_UNLOCK,
			Domain: d.metadata.Domain,
		},
	}
}

// GetAvailableStates implements LockObject.
func (d *lockObject) GetAvailableStates() []string {
	return []string{
		LOCK_STATE_JAMMED,
		LOCK_STATE_OPEN,
		LOCK_STATE_OPENING,
		LOCK_STATE_LOCKED,
		LOCK_STATE_LOCKING,
		LOCK_STATE_UNLOCKED,
		LOCK_STATE_UNLOCKING,
		LOCK_STATE_UNKNOWN,
	}
}

// GetMetadata implements LockObject.
func (d *lockObject) GetMetadata() ObjectMetadata {
	d.metadata.Type = "lock"
	return d.metadata
}

// RunAction implements LockObject.
func (d *lockObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	switch action {
	case LOCK_ACTION_LOCK:
		if d.lockMethod == nil {
			return nil, errors.New("lock method is not set")
		}
		return d.lockMethod(d, d.controller)
	case LOCK_ACTION_UNLOCK:
		if d.unlockMethod == nil {
			return nil, errors.New("unlock method is not set")
		}
		return d.unlockMethod(d, d.controller)
	}
	return nil, nil
}

// SetState implements LockObject.
func (d *lockObject) SetState(state string) error {
	return d.controller.SetState(d.metadata.ObjectID, state)
}

// Setup implements LockObject.
func (d *lockObject) Setup(controller ObjectController) error {
	d.controller = controller
	if d.setup == nil {
		return nil
	}
	return d.setup(d, controller)
}

// UpdateStateAttributes implements LockObject.
func (d *lockObject) UpdateStateAttributes(attributes map[string]string) error {
	return d.controller.UpdateStateAttributes(d.metadata.ObjectID, attributes)
}

type NewLockObjectParams struct {
	Metadata     ObjectMetadata
	LockMethod   func(this LockObject, controller ObjectController) (map[string]string, error)
	UnlockMethod func(this LockObject, controller ObjectController) (map[string]string, error)
	Setup        func(this LockObject, controller ObjectController) error
}

func NewLockObject(params NewLockObjectParams) LockObject {
	return &lockObject{
		metadata:     params.Metadata,
		lockMethod:   params.LockMethod,
		unlockMethod: params.UnlockMethod,
		setup:        params.Setup,
	}
}
