package objects

const SWITCH_STATE_OFF = "switch.state.off"
const SWITCH_STATE_ON = "switch.state.on"

const SWITCH_ACTION_TURN_ON = "switch.action.turn_on"
const SWITCH_ACTION_TURN_OFF = "switch.action.turn_off"

type SwitchObject interface {
	// base interface
	RegistrableObject
	// Set the state of the switch to on.
	TurnOn() error
	// Set the state of the switch to off.
	TurnOff() error
}
type switchObject struct {
	metadata      ObjectMetadata
	switchActions SwitchActions
	controller    ObjectController
}

// SetState implements SwitchObject.
func (s *switchObject) SetState(state string) error {
	return s.controller.SetState(s.metadata.ObjectID, state)

}

// AddEventTypes implements SwitchObject.
func (s *switchObject) AddEventTypes(eventTypes []EventType) error {
	panic("unimplemented")
}

// TurnOff implements SwitchObject.
func (s *switchObject) TurnOff() error {
	return s.controller.SetState(s.metadata.ObjectID, SWITCH_STATE_OFF)
}

// TurnOn implements SwitchObject.
func (s *switchObject) TurnOn() error {
	return s.controller.SetState(s.metadata.ObjectID, SWITCH_STATE_ON)
}

// GetMetadata implements RegistrableObject.
func (s *switchObject) GetMetadata() ObjectMetadata {
	s.metadata.Type = "switch"
	return s.metadata
}

// RunAction implements RegistrableObject.
func (s *switchObject) RunAction(action string, payload []byte) error {
	switch action {
	case SWITCH_ACTION_TURN_ON:
		return s.switchActions.TurnOn(s, s.controller)
	case SWITCH_ACTION_TURN_OFF:
		return s.switchActions.TurnOff(s, s.controller)
	}
	return nil
}

// GetAvailableActions implements RegistrableObject.
func (s *switchObject) GetAvailableActions() []ObjectAction {
	actionsresponse := []ObjectAction{}
	actions := []string{SWITCH_ACTION_TURN_ON, SWITCH_ACTION_TURN_OFF}

	for _, action := range actions {
		actionsresponse = append(actionsresponse, ObjectAction{
			Action: action,
			Domain: s.metadata.Domain,
		})
	}
	return actionsresponse
}

// GetAvailableStates implements RegistrableObject.
func (s *switchObject) GetAvailableStates() []string {
	return []string{SWITCH_STATE_OFF, SWITCH_STATE_ON}
}

// New implements RegistrableObject.
func (s *switchObject) Setup(oc ObjectController) error {
	s.controller = oc
	return s.switchActions.Setup(s, oc)
}

type SwitchActions struct {
	TurnOn  func(this RegistrableObject, oc ObjectController) error
	TurnOff func(this RegistrableObject, oc ObjectController) error
	Setup   func(this RegistrableObject, oc ObjectController) error
}

type NewSwitchObjectParams struct {
	Metadata      ObjectMetadata
	TurnOnMethod  func(this RegistrableObject, oc ObjectController) error
	TurnOffMethod func(this RegistrableObject, oc ObjectController) error
	SetupMethod   func(this RegistrableObject, oc ObjectController) error
}

func NewSwitchObject(params NewSwitchObjectParams) SwitchObject {

	return &switchObject{
		metadata: params.Metadata,
		switchActions: SwitchActions{
			TurnOn:  params.TurnOnMethod,
			TurnOff: params.TurnOffMethod,
			Setup:   params.SetupMethod,
		},
	}
}
