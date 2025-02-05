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
const READER_ACTION_STORE_FACES = "reader.action.store_faces"
const READER_ACTION_STORE_SMARTCARDS = "reader.action.store_smartcards"
const READER_ACTION_DELETE_QRS = "reader.action.delete_qrs"
const READER_ACTION_DELETE_FACES = "reader.action.delete_faces"
const READER_ACTION_DELETE_SMARTCARDS = "reader.action.delete_smartcards"

// domain
const READER_DOMAIN = "reader"

type QRsPayload struct {
	PersonId string   `json:"person_id"`
	Name     string   `json:"name"`
	Values   []string `json:"values"`
}

type FacesPayload struct {
	PersonId string   `json:"person_id"`
	Name     string   `json:"name"`
	Values   []string `json:"values"`
}

type SmartCardsPayload struct {
	PersonId string   `json:"person_id"`
	Name     string   `json:"name"`
	Values   []string `json:"values"`
}

type ReaderObject interface {
	RegistrableObject
}

type readerObject struct {
	metadata   ObjectMetadata
	controller ObjectController

	setupFunc func(this ReaderObject, controller ObjectController) error

	storeQRCredentials  func(this ReaderObject, controller ObjectController, payload QRsPayload) error
	deleteQRCredentials func(this ReaderObject, controller ObjectController, payload QRsPayload) error

	storeFaceCredentials  func(this ReaderObject, controller ObjectController, payload FacesPayload) error
	deleteFaceCredentials func(this ReaderObject, controller ObjectController, payload FacesPayload) error

	storeSmartCardCredentials  func(this ReaderObject, controller ObjectController, payload SmartCardsPayload) error
	deleteSmartCardCredentials func(this ReaderObject, controller ObjectController, payload SmartCardsPayload) error
}

// GetAvailableActions implements ReaderObject
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
			Action: READER_ACTION_STORE_FACES,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_STORE_SMARTCARDS,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_DELETE_QRS,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_DELETE_FACES,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_DELETE_SMARTCARDS,
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
		storeQrsPayload := QRsPayload{}
		if err := json.Unmarshal(payload, &storeQrsPayload); err != nil {
			return err
		}
		return r.storeQRCredentials(r, r.controller, storeQrsPayload)
	case READER_ACTION_DELETE_QRS:
		deleteQrsPayload := QRsPayload{}
		if err := json.Unmarshal(payload, &deleteQrsPayload); err != nil {
			return err
		}
		return r.deleteQRCredentials(r, r.controller, deleteQrsPayload)

	case READER_ACTION_STORE_FACES:
		storeFacesPayload := FacesPayload{}
		if err := json.Unmarshal(payload, &storeFacesPayload); err != nil {
			return err
		}
		return r.storeFaceCredentials(r, r.controller, storeFacesPayload)

	case READER_ACTION_DELETE_FACES:
		deleteFacesPayload := FacesPayload{}
		if err := json.Unmarshal(payload, &deleteFacesPayload); err != nil {
			return err
		}
		return r.deleteFaceCredentials(r, r.controller, deleteFacesPayload)

	case READER_ACTION_STORE_SMARTCARDS:
		storeSmartCardsPayload := SmartCardsPayload{}
		if err := json.Unmarshal(payload, &storeSmartCardsPayload); err != nil {
			return err
		}
		return r.storeSmartCardCredentials(r, r.controller, storeSmartCardsPayload)

	case READER_ACTION_DELETE_SMARTCARDS:
		deleteSmartCardsPayload := SmartCardsPayload{}
		if err := json.Unmarshal(payload, &deleteSmartCardsPayload); err != nil {
			return err
		}
		return r.deleteSmartCardCredentials(r, r.controller, deleteSmartCardsPayload)

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

	ReadMethod                 func(this ReaderObject, controller ObjectController) error
	StopMethod                 func(this ReaderObject, controller ObjectController) error
	ResetMethod                func(this ReaderObject, controller ObjectController) error
	RestartMethod              func(this ReaderObject, controller ObjectController) error
	StoreQRCredentials         func(this ReaderObject, controller ObjectController, payload QRsPayload) error
	DeleteQRCredentials        func(this ReaderObject, controller ObjectController, payload QRsPayload) error
	StoreFaceCredentials       func(this ReaderObject, controller ObjectController, payload FacesPayload) error
	DeleteFaceCredentials      func(this ReaderObject, controller ObjectController, payload FacesPayload) error
	StoreSmartCardCredentials  func(this ReaderObject, controller ObjectController, payload SmartCardsPayload) error
	DeleteSmartCardCredentials func(this ReaderObject, controller ObjectController, payload SmartCardsPayload) error
}

func NewReaderObject(params NewReaderObjectParams) ReaderObject {
	return &readerObject{
		metadata:                   params.Metadata,
		setupFunc:                  params.SetupFunc,
		storeQRCredentials:         params.StoreQRCredentials,
		deleteQRCredentials:        params.DeleteQRCredentials,
		storeFaceCredentials:       params.StoreFaceCredentials,
		deleteFaceCredentials:      params.DeleteFaceCredentials,
		storeSmartCardCredentials:  params.StoreSmartCardCredentials,
		deleteSmartCardCredentials: params.DeleteSmartCardCredentials,
	}
}
