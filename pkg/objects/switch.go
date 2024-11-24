package objects

const SWITCH_STATE_OFF = "switch.state.off"
const SWITCH_STATE_ON = "switch.state.on"

type switchObject struct {
	metadata      ObjectMetadata
	switchActions SwitchActions
	controller    ObjectController
}

// GetMetadata implements RegistrableObject.
func (s *switchObject) GetMetadata() ObjectMetadata {
	s.metadata.Type = "switch"
	return s.metadata
}

// SetMetadata implements RegistrableObject.
func (s *switchObject) SetMetadata(metadata ObjectMetadata) error {
	s.metadata = metadata
	return nil
}

// RunAction implements RegistrableObject.
func (s *switchObject) RunAction(action string, payload []byte) error {
	switch action {
	case "switch.action.turn_on":
		return s.switchActions.TurnOn(s, s.controller)
	case "switch.action.turn_off":
		return s.switchActions.TurnOff(s, s.controller)
	case "switch.action.toggle":
		return s.switchActions.Toggle(s, s.controller)
	}
	return nil
}

// GetAvailableActions implements RegistrableObject.
func (s *switchObject) GetAvailableActions() []ObjectAction {
	actionsresponse := []ObjectAction{}
	actions := []string{"switch.action.toggle", "switch.action.turn_on", "switch.action.turn_off"}

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
	return []string{"switch.state.off", "switch.state.on"}
}

// New implements RegistrableObject.
func (s *switchObject) Setup(oc ObjectController) error {
	s.controller = oc
	return s.switchActions.Setup(s, oc)
}

type SwitchActions struct {
	TurnOn  func(this RegistrableObject, oc ObjectController) error
	TurnOff func(this RegistrableObject, oc ObjectController) error
	Toggle  func(this RegistrableObject, oc ObjectController) error
	Setup   func(this RegistrableObject, oc ObjectController) error
}

func NewSwitchObject(objectMetadata ObjectMetadata, actions SwitchActions) (RegistrableObject, error) {
	if objectMetadata.ObjectID == "" {
		return nil, ErrObjectIdMandatory
	} else if objectMetadata.Name == "" {
		return nil, ErrNameMandatory
	} else if objectMetadata.Domain == "" {
		return nil, ErrDomainMandatory
	} else if objectMetadata.DeviceID == 0 {
		return nil, ErrDeviceIdMandatory
	} else if actions.TurnOn == nil || actions.TurnOff == nil || actions.Toggle == nil || actions.Setup == nil {
		return nil, ErrActionsMandatory
	}

	return &switchObject{
		metadata:      objectMetadata,
		switchActions: actions,
	}, nil
}
