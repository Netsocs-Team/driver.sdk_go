package objects

import (
	"fmt"

	"github.com/goccy/go-json"
)

type RelativeZoneObject interface {
	RegistrableObject
}

type RelativeZoneVertice struct {
	X float64
	Y float64
}

type RelativeZoneShape struct {
	Vertices []RelativeZoneVertice
}

type NewRelativeZoneObjectParams struct {
	Metadata ObjectMetadata
	Shape    RelativeZoneShape
	SetupFn  func(this RelativeZoneObject, controller ObjectController) error
}

type relativeZoneObject struct {
	setupFn    func(this RelativeZoneObject, controller ObjectController) error
	metadata   ObjectMetadata
	shape      RelativeZoneShape
	controller ObjectController
}

// GetAvailableActions implements RelativeZoneObject.
func (r *relativeZoneObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}
}

// GetAvailableStates implements RelativeZoneObject.
func (r *relativeZoneObject) GetAvailableStates() []string {
	return []string{}
}

// GetMetadata implements RelativeZoneObject.
func (r *relativeZoneObject) GetMetadata() ObjectMetadata {
	r.metadata.Type = "relative_zone"
	return r.metadata
}

// RunAction implements RelativeZoneObject.
func (r *relativeZoneObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	return nil, nil
}

// SetState implements RelativeZoneObject.
func (r *relativeZoneObject) SetState(state string) error {
	if r.controller == nil {
		return fmt.Errorf("controller is not set")
	}
	return r.controller.SetState(r.metadata.ObjectID, state)
}

// Setup implements RelativeZoneObject.
func (r *relativeZoneObject) Setup(controller ObjectController) error {
	r.controller = controller

	shape, err := json.Marshal(r.shape)
	if err != nil {
		return fmt.Errorf("failed to marshal shape: %w", err)
	}
	r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
		"shape": string(shape),
		"state": "active",
	})

	if r.setupFn != nil {
		return r.setupFn(r, controller)
	}
	return nil
}

// UpdateStateAttributes implements RelativeZoneObject.
func (r *relativeZoneObject) UpdateStateAttributes(attributes map[string]string) error {
	if r.controller == nil {
		return fmt.Errorf("controller is not set")
	}
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, attributes)
}

func NewRelativeZoneObject(params NewRelativeZoneObjectParams) RelativeZoneObject {
	return &relativeZoneObject{
		metadata: params.Metadata,
		shape:    params.Shape,
		setupFn:  params.SetupFn,
	}
}
