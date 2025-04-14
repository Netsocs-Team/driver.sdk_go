package objects

import (
	"fmt"

	"github.com/goccy/go-json"
)

const NOTIFIER_STATE_UNKNOWN = "notifier.state.unknown"
const NOTIFIER_STATE_IDLE = "notifier.state.idle"
const NOTIFIER_STATE_BUSY = "notifier.state.busy"
const NOTIFIER_STATE_ERROR = "notifier.state.error"
const NOTIFIER_ACTION_NOTIFY = "notifier.action.notify"

type NotifierObject interface {
	RegistrableObject
}

type NotificationPayload struct {
	Title    string `json:"title,omitempty"`
	Content  string `json:"content,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	AudioURL string `json:"audio_url,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
}

type notifierObject struct {
	controller ObjectController
	metadata   ObjectMetadata
	setupFn    func(notifierObject NotifierObject, oc ObjectController) error
	notifyFn   func(notifierObject NotifierObject, oc ObjectController, payload NotificationPayload) error
}

// GetAvailableActions implements NotifierObject.
func (n *notifierObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: NOTIFIER_ACTION_NOTIFY,
			Domain: n.GetMetadata().Domain,
		},
	}
}

// GetAvailableStates implements NotifierObject.
func (n *notifierObject) GetAvailableStates() []string {
	return []string{
		NOTIFIER_STATE_UNKNOWN,
		NOTIFIER_STATE_IDLE,
		NOTIFIER_STATE_BUSY,
		NOTIFIER_STATE_ERROR,
	}
}

// GetMetadata implements NotifierObject.
func (n *notifierObject) GetMetadata() ObjectMetadata {
	n.metadata.Type = "notifier"
	return n.metadata
}

// RunAction implements NotifierObject.
func (n *notifierObject) RunAction(id string, action string, payload []byte) (map[string]string, error) {
	var p NotificationPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	switch action {
	case NOTIFIER_ACTION_NOTIFY:
		if err := n.notifyFn(n, n.controller, p); err != nil {
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
	SetupFn  func(notifierObject NotifierObject, oc ObjectController) error
	NotifyFn func(notifierObject NotifierObject, oc ObjectController, payload NotificationPayload) error
}

func NewNotifierObject(props NewNotifierObjectProps) NotifierObject {
	n := &notifierObject{

		metadata: props.Metadata,
		setupFn:  props.SetupFn,
		notifyFn: props.NotifyFn,
	}

	return n
}
