package objects

import "fmt"

type VideoEngineObject interface {
	RegistrableObject
}

const VIDEO_ENGINE_DOMAIN = "video_engine"

type videoEngineObject struct {
	metadata   ObjectMetadata
	setup      func(this VideoEngineObject, controller ObjectController) error
	controller ObjectController
}

// UpdateStateAttributes implements VideoEngineObject.
func (v *videoEngineObject) UpdateStateAttributes(attributes map[string]string) error {
	return v.controller.UpdateStateAttributes(v.metadata.ObjectID, attributes)
}

// SetState implements VideoEngineObject.
func (v *videoEngineObject) SetState(state string) error {
	return v.controller.SetState(v.metadata.ObjectID, state)
}

// GetAvailableActions implements VideoEngineObject.
func (v *videoEngineObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}
}

// GetAvailableStates implements VideoEngineObject.
func (v *videoEngineObject) GetAvailableStates() []string {
	return []string{"online", "offline", "error"}
}

// GetMetadata implements VideoEngineObject.
func (v *videoEngineObject) GetMetadata() ObjectMetadata {
	v.metadata.Type = VIDEO_ENGINE_DOMAIN
	return v.metadata
}

// RunAction implements VideoEngineObject.
func (v *videoEngineObject) RunAction(id, action string, payload []byte) (map[string]string, error) {

	return nil, fmt.Errorf("action %s not found", action)
}

// Setup implements VideoEngineObject.
func (v *videoEngineObject) Setup(oc ObjectController) error {
	v.controller = oc
	return v.setup(v, oc)
}

type NewVideoEngineObjectParams struct {
	Metadata ObjectMetadata
	Setup    func(this VideoEngineObject, controller ObjectController) error
}

func NewVideoEngineObject(params NewVideoEngineObjectParams) VideoEngineObject {
	return &videoEngineObject{
		metadata: params.Metadata,
		setup:    params.Setup,
	}
}
