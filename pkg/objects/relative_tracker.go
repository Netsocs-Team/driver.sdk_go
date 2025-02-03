package objects

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
func (r *relativeTrackerObject) RunAction(action string, payload []byte) error {
	return nil
}

// SetAcceleration implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetAcceleration(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]interface{}{
		"acceleration_x": x,
		"acceleration_y": y,
		"acceleration_z": z,
	})
}

// SetCoords implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetCoords(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]interface{}{
		"position_x": x,
		"position_y": y,
		"position_z": z,
	})
}

// SetSize implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetSize(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]interface{}{
		"size_x": x,
		"size_y": y,
		"size_z": z,
	})
}

// SetState implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetState(state string) error {
	return r.controller.SetState(r.metadata.ObjectID, state)
}

// SetVelocity implements RelativeTrackerObject.
func (r *relativeTrackerObject) SetVelocity(x float64, y float64, z float64) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, map[string]interface{}{
		"velocity_x": x,
		"velocity_y": y,
		"velocity_z": z,
	})
}

// Setup implements RelativeTrackerObject.
func (r *relativeTrackerObject) Setup(controller ObjectController) error {
	r.controller = controller
	return nil
}

type NewRelativeTrackerObjectProps struct {
	Metadata ObjectMetadata
}

func NewRelativeTrackerObject(props NewRelativeTrackerObjectProps) RelativeTrackerObject {
	return &relativeTrackerObject{
		metadata: props.Metadata,
	}
}
