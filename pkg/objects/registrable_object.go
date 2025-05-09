package objects

type EventType struct {
	Domain             string `json:"domain"`
	DisplayName        string `json:"display_name"`
	DisplayDescription string `json:"display_description"`
	EventType          string `json:"event_type"`
	EventLevel         string `json:"event_level"`
	Color              string `json:"color"`
	ShowColor          bool   `json:"show_color"`
	IsHidden           bool   `json:"is_hidden"`
	Origin             string `json:"origin"`
}

type RegistrableObject interface {
	Setup(ObjectController) error
	GetAvailableStates() []string
	GetAvailableActions() []ObjectAction
	RunAction(id, action string, payload []byte) (map[string]string, error)
	GetMetadata() ObjectMetadata
	SetState(state string) error
	UpdateStateAttributes(attributes map[string]string) error
}
