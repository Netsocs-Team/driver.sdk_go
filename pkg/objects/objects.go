package objects

type ObjectRunner interface {
	RegisterObject(object RegistrableObject) error
}

type ObjectController interface {
	SetState(objectId string, state string) error
	NewAction(action ObjectAction) error
	CreateObject(RegistrableObject) error
	ListenActionRequests() error
}

type ObjectMetadata struct {
	ObjectID string            `json:"object_id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Domain   string            `json:"domain"`
	I18n     map[string]string `json:"i18n"`
	DeviceID int               `json:"device_id"`
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
	SetMetadata(metadata ObjectMetadata) error
	GetMetadata() ObjectMetadata
}
