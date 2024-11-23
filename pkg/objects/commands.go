package objects

type Command string

const (
	/*
		Say to driverhub that need change attribute of object.
		Driverhub will save a record that this request and the DriverHandler will receive this request and will change the attribute of the object.
		Then the same object will receive and call the method SetAttribute.

		The params to used is a JSON string with the attribute name and the value to set.
		Example: `{"key":"isOn","value":"true"}`
		Where key is the attribute name and value is the value to set.
	*/
	_SET_STATE_COMMAND Command = "change_attribute"
)

type changeAttributeCommandParams struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type commandToObjectHandler struct {
	ObjectID string  `json:"object_id"`
	Command  Command `json:"command"`
	Params   string  `json:"params"`
}
