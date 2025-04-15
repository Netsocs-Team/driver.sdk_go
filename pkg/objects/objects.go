package objects

type ObjectRunner interface {
	RegisterObject(object RegistrableObject) error
	GetController() ObjectController
}

type SetupFunction func(RegistrableObject, ObjectController) error

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
