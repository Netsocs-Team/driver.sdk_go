package objects

type ObjectRunner interface {
	RegisterObject(object RegistrableObject) error
}

type SetupFunction func(RegistrableObject, ObjectController) error

type ObjectController interface {
	SetState(objectId string, state string) error
	UpdateStateAttributes(objectId string, attributes map[string]interface{}) error
	NewAction(action ObjectAction) error
	CreateObject(RegistrableObject) error
	ListenActionRequests() error
	GetDriverhubHost() string
	GetDriverKey() string
}

type ObjectMetadata struct {
	ObjectID string            `json:"object_id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Domain   string            `json:"domain"`
	I18n     map[string]string `json:"i18n"`
	DeviceID string            `json:"device_id"`
	Tags     []string          `json:"tags"`
}

type ObjectAction struct {
	Action string `json:"action"`
	Domain string `json:"domain"`
}

type RegistrableObject interface {
	Setup(ObjectController) error
	GetAvailableStates() []string
	GetAvailableActions() []ObjectAction
	RunAction(action string, payload []byte) error
	GetMetadata() ObjectMetadata
}
