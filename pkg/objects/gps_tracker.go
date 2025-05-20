package objects

type GpsTrackerObject interface {
	RegistrableObject
}

type gpsTrackerObject struct {
	metadata   ObjectMetadata
	controller ObjectController
	setupFn    func(GpsTrackerObject, ObjectController) error
}

// GetAvailableActions implements GpsTrackerObject.
func (g *gpsTrackerObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}

}

// GetAvailableStates implements GpsTrackerObject.
func (g *gpsTrackerObject) GetAvailableStates() []string {
	return []string{}
}

// GetMetadata implements GpsTrackerObject.
func (g *gpsTrackerObject) GetMetadata() ObjectMetadata {
	g.metadata.Type = "gps_tracker"
	return g.metadata
}

// RunAction implements GpsTrackerObject.
func (g *gpsTrackerObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	return nil, nil
}

// SetState implements GpsTrackerObject.
func (g *gpsTrackerObject) SetState(state string) error {
	return g.controller.SetState(g.GetMetadata().ObjectID, state)
}

// Setup implements GpsTrackerObject.
func (g *gpsTrackerObject) Setup(oc ObjectController) error {
	g.controller = oc
	if g.setupFn != nil {
		return g.setupFn(g, oc)
	}
	return nil
}

// UpdateStateAttributes implements GpsTrackerObject.
func (g *gpsTrackerObject) UpdateStateAttributes(attributes map[string]string) error {
	return g.controller.UpdateStateAttributes(g.GetMetadata().ObjectID, attributes)
}

type NewGPSTrackerObjectProps struct {
	Metadata ObjectMetadata
	// SetupFn is a function that will be called to setup the object
	SetupFn func(GpsTrackerObject, ObjectController) error
}

func NewGPSTrackerObject(props NewGPSTrackerObjectProps) GpsTrackerObject {
	return &gpsTrackerObject{
		metadata: props.Metadata,
		setupFn:  props.SetupFn,
	}
}
