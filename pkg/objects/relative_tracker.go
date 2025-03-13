package objects

import "strconv"

const RELATIVE_TRACKER_STATE_MOVING = "relative_tracker.state.moving"
const RELATIVE_TRACKER_STATE_NO_SIGNAL = "relative_tracker.state.no_signal"

type RelativeTrackerObject interface {
	RegistrableObject

	SetCoords(x, y, z float64) error
	SetVelocity(x, y, z float64) error
	SetAcceleration(x, y, z float64) error
	SetSize(x, y, z float64) error

	SetMovimingState() error
	SetNoSignalState() error
}

type relativeTrackerObject struct {
	metadata   ObjectMetadata
	controller ObjectController

	setupFn func(this RelativeTrackerObject, controller ObjectController) error
}

// UpdateStateAttributes implements RelativeTrackerObject.
func (r *relativeTrackerObject) UpdateStateAttributes(attributes map[string]string) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, attributes)
}

// SetMovimingState implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetMovimingState() error {
	return r.SetState(RELATIVE_TRACKER_STATE_MOVING)
}

// SetNoSignalState implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetNoSignalState() error {
	return r.SetState(RELATIVE_TRACKER_STATE_NO_SIGNAL)
}

// GetAvailableActions implements RelativeTrackerObject.
func (r *relativeTrackerObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}
}

// GetAvailableStates implements RelativeTrackerObject.
func (r *relativeTrackerObject) GetAvailableStates() []string {
	return []string{RELATIVE_TRACKER_STATE_MOVING, RELATIVE_TRACKER_STATE_NO_SIGNAL}
}

// GetMetadata implements RelativeTrackerObject.
func (r *relativeTrackerObject) GetMetadata() ObjectMetadata {
	r.metadata.Type = "relative_tracker"
	return r.metadata
}

// RunAction implements RelativeTrackerObject.
func (r *relativeTrackerObject) RunAction(id, action string, payload []byte) error {
	return nil
}

// SetAcceleration implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetAcceleration(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
		"acceleration_x": strconv.FormatFloat(x, 'f', -1, 64),
		"acceleration_y": strconv.FormatFloat(y, 'f', -1, 64),
		"acceleration_z": strconv.FormatFloat(z, 'f', -1, 64),
	})
}

// SetCoords implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetCoords(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
		"position_x": strconv.FormatFloat(x, 'f', -1, 64),
		"position_y": strconv.FormatFloat(y, 'f', -1, 64),
		"position_z": strconv.FormatFloat(z, 'f', -1, 64),
	})
}

// SetSize implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetSize(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
		"size_x": strconv.FormatFloat(x, 'f', -1, 64),
		"size_y": strconv.FormatFloat(y, 'f', -1, 64),
		"size_z": strconv.FormatFloat(z, 'f', -1, 64),
	})
}

// SetState implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetState(state string) error {
	return r.controller.SetState(r.metadata.ObjectID, state)
}

// SetVelocity implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetVelocity(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
		"velocity_x": strconv.FormatFloat(x, 'f', -1, 64),
		"velocity_y": strconv.FormatFloat(y, 'f', -1, 64),
		"velocity_z": strconv.FormatFloat(z, 'f', -1, 64),
	})
}

// Setup implements RelativeTrackerObject.
func (r *relativeTrackerObject) Setup(controller ObjectController) error {
	r.controller = controller
	if r.setupFn != nil {
		return r.setupFn(r, controller)
	}
	return nil
}

type NewRelativeTrackerObjectProps struct {
	Metadata ObjectMetadata
	SetupFn  func(this RelativeTrackerObject, controller ObjectController) error
}

func NewRelativeTrackerObject(props NewRelativeTrackerObjectProps) RelativeTrackerObject {
	return &relativeTrackerObject{
		metadata: props.Metadata,
		setupFn:  props.SetupFn,
	}
}
