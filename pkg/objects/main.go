package objects

// The object is an abstraction of a component of the application
// which has states, methods, and properties that define it.
// The developer can create objects of this type to represent
// functionalities of the integration and unify in a language that can
// be understood by the Netsocs ecosystem.
type Object struct {
	// The ID of an object is constructed with a format of driver_name:device_id:descriptive_name
	// for example: "netsocs_hardware:123456:relay_1"
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	// The State of an object is a string that represents the current state of the object.
	State           string            `json:"state"`
	DeviceID        int               `json:"deviceID"`
	StateProperties map[string]string `json:"stateProperties"`
	Enabled         bool              `json:"enabled"`
	Icon            string            `json:"icon"`
	// The object runner can use this channel to send commands to the object handler
	// that will be responsible for executing the commands.
	// Its is utilized by the object runner to send commands until the DriverHub.
	CommunicationToHandlerChannel chan commandToObjectHandler
}

func (o *Object) SetState(state string) {
	o.State = state
}

func (o *Object) SetStateProperties(key, value string) {
	if o.StateProperties == nil {
		o.StateProperties = make(map[string]string)
	}
	o.StateProperties[key] = value
}

func (o *Object) SetIcon(icon string) {
	o.Icon = icon
}

func (o *Object) GetState() string {
	return o.State
}

func (o *Object) GetStateProperties(key string) string {
	return o.StateProperties[key]
}

func (o *Object) GetIcon() string {
	return o.Icon
}
