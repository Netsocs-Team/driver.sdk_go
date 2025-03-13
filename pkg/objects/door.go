package objects

type DoorObject interface {
	RegistrableObject
}

type doorObject struct {
	metadata        ObjectMetadata
	openDoorMethod  func(this DoorObject, controller ObjectController) error
	closeDoorMethod func(this DoorObject, controller ObjectController) error
	setup           func(this DoorObject, controller ObjectController) error
	controller      ObjectController
}

// UpdateStateAttributes implements DoorObject.
func (d *doorObject) UpdateStateAttributes(attributes map[string]string) error {
	return d.controller.UpdateStateAttributes(d.metadata.ObjectID, attributes)
}

// SetState implements DoorObject.
func (d *doorObject) SetState(state string) error {
	return d.controller.SetState(d.metadata.ObjectID, state)
}

const DOOR_STATE_OPEN = "door.state.open"
const DOOR_STATE_CLOSE = "door.state.close"
const DOOR_STATE_LOCK = "door.state.lock"
const DOOR_STATE_OPENING = "door.state.opening"
const DOOR_STATE_CLOSING = "door.state.closing"
const DOOR_STATE_UNKNOWN = "door.state.unknown"

const DOOR_ACTION_OPEN = "door.action.open"
const DOOR_ACTION_CLOSE = "door.action.close"

const DOOR_DOMAIN = "door"

// GetAvailableActions implements DoorObject.
func (d *doorObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: DOOR_ACTION_OPEN,
			Domain: d.metadata.Domain,
		},
		{
			Action: DOOR_ACTION_CLOSE,
			Domain: d.metadata.Domain,
		},
	}
}

// GetAvailableStates implements DoorObject.
func (d *doorObject) GetAvailableStates() []string {
	return []string{DOOR_STATE_CLOSE, DOOR_STATE_CLOSING, DOOR_STATE_LOCK, DOOR_STATE_OPEN, DOOR_STATE_OPENING}
}

// GetMetadata implements DoorObject.
func (d *doorObject) GetMetadata() ObjectMetadata {
	d.metadata.Type = DOOR_DOMAIN
	return d.metadata
}

// RunAction implements DoorObject.
func (d *doorObject) RunAction(id, action string, payload []byte) error {
	switch action {
	case DOOR_ACTION_OPEN:
		if err := d.openDoorMethod(d, d.controller); err != nil {
			return err
		}
		return d.controller.SetState(d.metadata.ObjectID, DOOR_STATE_OPEN)
	case DOOR_ACTION_CLOSE:
		if err := d.closeDoorMethod(d, d.controller); err != nil {
			return err
		}
		return d.controller.SetState(d.metadata.ObjectID, DOOR_STATE_CLOSE)
	}
	return nil
}

// Setup implements DoorObject.
func (d *doorObject) Setup(oc ObjectController) error {
	d.controller = oc

	return d.setup(d, oc)
}

type NewDoorObjectParams struct {
	Metadata        ObjectMetadata
	OpenDoorMethod  func(this DoorObject, controller ObjectController) error
	CloseDoorMethod func(this DoorObject, controller ObjectController) error
	Setup           func(this DoorObject, controller ObjectController) error
}

func NewDoorObject(params NewDoorObjectParams) DoorObject {

	return &doorObject{
		metadata:        params.Metadata,
		openDoorMethod:  params.OpenDoorMethod,
		closeDoorMethod: params.CloseDoorMethod,
		setup:           params.Setup,
	}
}
