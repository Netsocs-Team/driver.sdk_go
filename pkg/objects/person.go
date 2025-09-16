package objects

type personObject struct {
	metadata   ObjectMetadata
	controller ObjectController
}

// GetAvailableActions implements RegistrableObject.
func (p *personObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}
}

// GetAvailableStates implements RegistrableObject.
func (p *personObject) GetAvailableStates() []string {
	return []string{}
}

// GetMetadata implements RegistrableObject.
func (p *personObject) GetMetadata() ObjectMetadata {
	return p.metadata
}

// RunAction implements RegistrableObject.
func (p *personObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	return nil, nil
}

// SetState implements RegistrableObject.
func (p *personObject) SetState(state string) error {
	return p.controller.SetState(p.metadata.ObjectID, state)
}

// Setup implements RegistrableObject.
func (p *personObject) Setup(oc ObjectController) error {
	p.controller = oc
	return nil
}

// UpdateStateAttributes implements RegistrableObject.
func (p *personObject) UpdateStateAttributes(attributes map[string]string) error {
	return p.controller.UpdateStateAttributes(p.metadata.ObjectID, attributes)
}

type NewPersonObjectParams struct {
	Metadata ObjectMetadata
}

func NewPersonObject(params NewPersonObjectParams) RegistrableObject {
	return &personObject{
		metadata: params.Metadata,
	}
}
