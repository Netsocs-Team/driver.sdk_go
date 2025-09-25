package objects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/goccy/go-json"
)

// states
const READER_STATE_READING = "reader.state.reading"
const READER_STATE_IDLE = "reader.state.idle"
const READER_STATE_UNKNOWN = "reader.state.unknown"
const READER_STATE_ERROR = "reader.state.error"

// actions
const READER_ACTION_READ = "read"
const READER_ACTION_STOP = "reader.action.stop"
const READER_ACTION_RESET = "reader.action.reset"
const READER_ACTION_RESTART = "reader.action.restart"
const READER_ACTION_STORE_QRS = "reader.action.store_qrs"
const READER_ACTION_DELETE_QRS = "reader.action.delete_qrs"
const READER_ACTION_DELETE_PERSON = "reader.action.delete_person"
const READER_ACTION_GET_PEOPLE = "get_people"
const READER_ACTION_SET_PEOPLE = "set_people"

// domain
const READER_DOMAIN = "reader"

type DeletePersonPayload struct {
	PersonId string `json:"person_id"`
}

type QRPayload struct {
	PersonId string   `json:"person_id"`
	Name     string   `json:"name"`
	Values   []string `json:"values"`
}

type ReaderObject interface {
	RegistrableObject
}

type ReaderPeople struct {
	People          []ReaderPerson         `json:"people"`
	SupportSchedule bool                   `json:"support_schedule"`
	Schedules       []ReaderPersonSchedule `json:"schedule"`
}

type ReaderPersonSchedule struct {
	LastUpdated string                        `json:"last_updated"`
	ID          string                        `json:"id"`
	Monday      ReaderPersonScheduleDay       `json:"monday"`
	Tuesday     ReaderPersonScheduleDay       `json:"tuesday"`
	Wednesday   ReaderPersonScheduleDay       `json:"wednesday"`
	Thursday    ReaderPersonScheduleDay       `json:"thursday"`
	Friday      ReaderPersonScheduleDay       `json:"friday"`
	Saturday    ReaderPersonScheduleDay       `json:"saturday"`
	Sunday      ReaderPersonScheduleDay       `json:"sunday"`
	Holidays    []ReaderPersonScheduleHoliday `json:"holidays"`
}

type ReaderPersonScheduleHoliday struct {
	Date    string `json:"date"`
	Enabled bool   `json:"enabled"`
}

type ReaderPersonScheduleDay struct {
	Start   string `json:"start"` // RFC 3339
	End     string `json:"end"`   // RFC 3339
	Enabled bool   `json:"enabled"`
}

type ReaderPerson struct {
	PersonId    string             `json:"person_id"`
	Name        string             `json:"name"`
	Credentials []ReaderCredential `json:"credentials"`
}

type ReaderCredential struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	LastUpdated string `json:"last_updated"`
}

type CredentialType string

const (
	CREDENTIAL_TYPE_FACE                      CredentialType = "face"
	CREDENTIAL_TYPE_NORMAL_CARD               CredentialType = "normal_card"
	CREDENTIAL_TYPE_FINGERPRINT_ISO_19794_2   CredentialType = "fingerprint_iso_19794_2"
	CREDENTIAL_TYPE_FINGERPRINT_ANSI_378_2004 CredentialType = "fingerprint_ansi_378_2004"
)

type ReadCreadentialPayload struct {
	Type CredentialType `json:"type"`
}

type ReadCredentialResponse struct {
	Data     string            `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type readerObject struct {
	metadata   ObjectMetadata
	controller ObjectController

	setupFunc                func(this ReaderObject, controller ObjectController) error
	restart                  func(this ReaderObject, controller ObjectController) error
	storeQRCredentials       func(this ReaderObject, controller ObjectController, payload QRPayload) error
	deleteQRCredentials      func(this ReaderObject, controller ObjectController, payload QRPayload) error
	deletePersonCredentials  func(this ReaderObject, controller ObjectController, payload DeletePersonPayload) error
	getPeopleCredentials     func(this ReaderObject, controller ObjectController) (ReaderPeople, error)
	setPeopleCredentials     func(this ReaderObject, controller ObjectController, payload ReaderPeople) error
	storePeopleCredentials   func(this ReaderObject, controller ObjectController, payload ReaderPeople) error
	readCredential           func(this ReaderObject, controller ObjectController, payload ReadCreadentialPayload) (ReadCredentialResponse, error)
	supportedCredentialTypes []CredentialType
}

// UpdateStateAttributes implements ReaderObject.
func (r *readerObject) UpdateStateAttributes(attributes map[string]string) error {
	return r.controller.UpdateStateAttributes(r.metadata.ObjectID, attributes)
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
		{
			Action: READER_ACTION_DELETE_PERSON,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_GET_PEOPLE,
			Domain: r.metadata.Domain,
		},
		{
			Action: READER_ACTION_SET_PEOPLE,
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
func (r *readerObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
	switch action {

	case READER_ACTION_RESTART:
		return nil, r.restart(r, r.controller)

	case READER_ACTION_STORE_QRS:
		storeQrsPayload := QRPayload{}
		if err := json.Unmarshal(payload, &storeQrsPayload); err != nil {
			return nil, err
		}
		return nil, r.storeQRCredentials(r, r.controller, storeQrsPayload)

	case READER_ACTION_DELETE_QRS:
		deleteQrsPayload := QRPayload{}
		if err := json.Unmarshal(payload, &deleteQrsPayload); err != nil {
			return nil, err
		}
		return nil, r.deleteQRCredentials(r, r.controller, deleteQrsPayload)

	case READER_ACTION_DELETE_PERSON:
		deletePersonPayload := DeletePersonPayload{}
		if err := json.Unmarshal(payload, &deletePersonPayload); err != nil {
			return nil, err
		}
		return nil, r.deletePersonCredentials(r, r.controller, deletePersonPayload)

	case READER_ACTION_GET_PEOPLE:
		people, err := r.getPeopleCredentials(r, r.controller)
		if err != nil {
			return nil, err
		}
		peopleBytes, err := json.Marshal(people)
		if err != nil {
			return nil, err
		}
		return map[string]string{"people": string(peopleBytes)}, nil

	case READER_ACTION_SET_PEOPLE:
		people := ReaderPeople{}
		if err := json.Unmarshal(payload, &people); err != nil {
			return nil, err
		}
		return nil, r.setPeopleCredentials(r, r.controller, people)
	case READER_ACTION_READ:
		if r.readCredential == nil {
			return nil, fmt.Errorf("read credential method not implemented")
		}
		readCredentialPayload := ReadCreadentialPayload{}
		if err := json.Unmarshal(payload, &readCredentialPayload); err != nil {
			return nil, err
		}
		if !slices.Contains(r.supportedCredentialTypes, readCredentialPayload.Type) {
			return nil, fmt.Errorf("credential type '%s' not supported", readCredentialPayload.Type)
		}
		response, err := r.readCredential(r, r.controller, readCredentialPayload)
		if err != nil {
			return nil, err
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		return map[string]string{"data": string(responseBytes)}, nil
	}

	return nil, fmt.Errorf("action %s not found", action)
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

	if len(r.supportedCredentialTypes) > 0 {
		list := []string{}
		for _, t := range r.supportedCredentialTypes {
			list = append(list, string(t))
		}
		oc.UpdateStateAttributes(r.metadata.ObjectID, map[string]string{
			"supported_credential_types": strings.Join(list, ","),
		})
	}

	return r.setupFunc(r, oc)
}

type NewReaderObjectParams struct {
	SetupFunc func(this ReaderObject, controller ObjectController) error
	Metadata  ObjectMetadata

	ReadMethod                    func(this ReaderObject, controller ObjectController) error
	StopMethod                    func(this ReaderObject, controller ObjectController) error
	ResetMethod                   func(this ReaderObject, controller ObjectController) error
	RestartMethod                 func(this ReaderObject, controller ObjectController) error
	StoreQRCredentialsMethod      func(this ReaderObject, controller ObjectController, payload QRPayload) error
	DeleteQRCredentialsMethod     func(this ReaderObject, controller ObjectController, payload QRPayload) error
	DeletePersonCredentialsMethod func(this ReaderObject, controller ObjectController, payload DeletePersonPayload) error
	GetPeopleCredentialsMethod    func(this ReaderObject, controller ObjectController) (ReaderPeople, error)
	SetPeopleCredentialsMethod    func(this ReaderObject, controller ObjectController, payload ReaderPeople) error
	StorePeopleCredentialsMethod  func(this ReaderObject, controller ObjectController, payload ReaderPeople) error
	ReadCredentialMethod          func(this ReaderObject, controller ObjectController, payload ReadCreadentialPayload) (ReadCredentialResponse, error)
	SupportedCredentialTypes      []CredentialType
}

func NewReaderObject(params NewReaderObjectParams) ReaderObject {
	return &readerObject{
		metadata:                 params.Metadata,
		setupFunc:                params.SetupFunc,
		restart:                  params.RestartMethod,
		storeQRCredentials:       params.StoreQRCredentialsMethod,
		deleteQRCredentials:      params.DeleteQRCredentialsMethod,
		deletePersonCredentials:  params.DeletePersonCredentialsMethod,
		getPeopleCredentials:     params.GetPeopleCredentialsMethod,
		setPeopleCredentials:     params.SetPeopleCredentialsMethod,
		storePeopleCredentials:   params.StorePeopleCredentialsMethod,
		supportedCredentialTypes: params.SupportedCredentialTypes,
		readCredential:           params.ReadCredentialMethod,
	}
}
