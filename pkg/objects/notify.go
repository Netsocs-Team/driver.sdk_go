package objects

import (
	"fmt"

	"github.com/goccy/go-json"
)

const NOTIFY_STATE_UNKNOWN = "notify.state.unknown"
const NOTIFY_STATE_IDLE = "notify.state.idle"
const NOTIFY_STATE_BUSY = "notify.state.busy"
const NOTIFY_STATE_ERROR = "notify.state.error"
const CREATE = "create"

type NotifyObject interface {
	RegistrableObject
}

type CreatePayload struct {
	Message        string `json:"message"`
	Title          string `json:"title,omitempty"`
	NotificationID string `json:"notification_id,omitempty"`
	Target         string `json:"target,omitempty"`
	Data           struct {
		ImageURLs []string `json:"image_urls,omitempty"`
		AudioURLs []string `json:"audio_urls,omitempty"`
		VideoURLs []string `json:"video_urls,omitempty"`
	} `json:"data,omitempty"`
}

type notifierObject struct {
	controller ObjectController
	metadata   ObjectMetadata
	setupFn    func(notifierObject NotifyObject, oc ObjectController) error
	createFn   func(notifierObject NotifyObject, oc ObjectController, payload CreatePayload) error
}

// GetAvailableActions implements NotifierObject.
func (n *notifierObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: CREATE,
			Domain: n.GetMetadata().Domain,
		},
	}
}

// GetAvailableStates implements NotifierObject.
func (n *notifierObject) GetAvailableStates() []string {
	return []string{
		NOTIFY_STATE_UNKNOWN,
		NOTIFY_STATE_IDLE,
		NOTIFY_STATE_BUSY,
		NOTIFY_STATE_ERROR,
	}
}

// GetMetadata implements NotifierObject.
func (n *notifierObject) GetMetadata() ObjectMetadata {
	n.metadata.Type = "notifier"
	return n.metadata
}

// RunAction implements NotifierObject.
func (n *notifierObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	var p CreatePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	switch action {
	case CREATE:
		if err := n.createFn(n, n.controller, p); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, fmt.Errorf("action %s not found", action)
}

// SetState implements NotifierObject.
func (n *notifierObject) SetState(state string) error {
	return n.controller.SetState(n.GetMetadata().ObjectID, state)
}

// Setup implements NotifierObject.
func (n *notifierObject) Setup(oc ObjectController) error {
	n.controller = oc
	if n.setupFn != nil {
		return n.setupFn(n, oc)
	}
	return nil
}

// UpdateStateAttributes implements NotifierObject.
func (n *notifierObject) UpdateStateAttributes(attributes map[string]string) error {
	return n.controller.UpdateStateAttributes(n.GetMetadata().ObjectID, attributes)
}

type NewNotifierObjectProps struct {
	Metadata ObjectMetadata
	SetupFn  func(notifierObject NotifyObject, oc ObjectController) error
	CreateFn func(notifierObject NotifyObject, oc ObjectController, payload CreatePayload) error
}

func NewNotifierObject(props NewNotifierObjectProps) NotifyObject {
	n := &notifierObject{

		metadata: props.Metadata,
		setupFn:  props.SetupFn,
		createFn: props.CreateFn,
	}

	return n
}
