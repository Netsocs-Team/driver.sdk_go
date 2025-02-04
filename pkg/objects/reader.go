package objects

import "github.com/goccy/go-json"

// states
const READER_STATE_READING = "reader.state.reading"
const READER_STATE_IDLE = "reader.state.idle"
const READER_STATE_UNKNOWN = "reader.state.unknown"
const READER_STATE_ERROR = "reader.state.error"

// actions
const READER_ACTION_READ = "reader.action.read"
const READER_ACTION_STOP = "reader.action.stop"
const READER_ACTION_RESET = "reader.action.reset"
const READER_ACTION_RESTART = "reader.action.restart"
const READER_ACTION_STORE_QRS = "reader.action.store_qrs"
const READER_ACTION_DELETE_QRS = "reader.action.delete_qrs"

// domain
const READER_DOMAIN = "reader"

type PersonData struct {
	PersonId string `json:"personId"`
	Name     string `json:"name"`
}

type StoreQRsPayload struct {
	PersonData
	Values []string `json:"values"`
}

type ReaderObject interface {
	RegistrableObject
}

type readerObject struct {
	metadata   ObjectMetadata
	controller ObjectController

	setupFunc        func(this ReaderObject, controller ObjectController) error
	storeCredential  func(this ReaderObject, controller ObjectController, payload StoreQRsPayload) error
	deleteCredential func(this ReaderObject, controller ObjectController, payload StoreQRsPayload) error
}

// GetAvailableActions implements ReaderObject.
func (r *readerObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: READER_ACTION_READ,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_STOP,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_RESET,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_RESTART,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_STORE_QRS,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_DELETE_QRS,
			Domain: r.metadata.Domain,
		},
	}
}

// GetAvailableStates implements ReaderObject.
func (r *readerObject) GetAvailableStates() []string {
	return []string{READER_STATE_UNKNOWN, READER_STATE_IDLE, READER_STATE_READING, READER_STATE_ERROR}
}

// GetMetadata implements ReaderObject.
func (r *readerObject) GetMetadata() ObjectMetadata {
	r.metadata.Type = READER_DOMAIN
	return r.metadata
}

// RunAction implements ReaderObject.
func (r *readerObject) RunAction(action string, payload []byte) error {
	switch action {
	case READER_ACTION_STORE_QRS:
		storeQrsPayload := StoreQRsPayload{}
		if err := json.Unmarshal(payload, &payload); err != nil {
			return err
		}
		return r.storeCredential(r, r.controller, storeQrsPayload)
	}

	return nil
}

// SetState implements ReaderObject.
func (r *readerObject) SetState(state string) error {
	return r.controller.SetState(r.metadata.ObjectID, state)
}

// Setup implements ReaderObject.
func (r *readerObject) Setup(oc ObjectController) error {
	r.controller = oc

	if r.setupFunc == nil {
		return nil
	}

	return r.setupFunc(r, oc)
}

type NewReaderObjectParams struct {
	SetupFunc func(this ReaderObject, controller ObjectController) error
	Metadata  ObjectMetadata

	ReadMethod             func(this ReaderObject, controller ObjectController) error
	StopMethod             func(this ReaderObject, controller ObjectController) error
	ResetMethod            func(this ReaderObject, controller ObjectController) error
	RestartMethod          func(this ReaderObject, controller ObjectController) error
	StoreCredentialMethod  func(this ReaderObject, controller ObjectController, payload StoreQRsPayload) error
	DeleteCredentialMethod func(this ReaderObject, controller ObjectController, payload StoreQRsPayload) error
}

func NewReaderObject(params NewReaderObjectParams) ReaderObject {
	return &readerObject{
		metadata:         params.Metadata,
		setupFunc:        params.SetupFunc,
		storeCredential:  params.StoreCredentialMethod,
		deleteCredential: params.DeleteCredentialMethod,
	}
}
