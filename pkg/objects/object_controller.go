package objects

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/logger"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/tools"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type ObjectController interface {
	SetState(objectId string, state string) error
	UpdateStateAttributes(objectId string, attributes map[string]string) error
	UpdateResultAttributes(ActionExecutionID string, attributes map[string]string) error
	NewAction(action ObjectAction) error
	CreateObject(RegistrableObject) error
	ListenActionRequests() error
	GetDriverhubHost() string
	GetDriverKey() string
	GetState(objectId string) (state StateRecord, err error)
	DisabledObject(objectId string) error
	EnabledObject(objectId string) error
	AddEventTypes(eventTypes []EventType) error
	Increment(objectId string) error
	Decrement(objectId string) error
}

type objectController struct {
	driverhub_host string
	driver_key     string
	token          string
	httpClient     *resty.Client
}

// GetState implements ObjectController.
func (o *objectController) GetState(objectId string) (state StateRecord, err error) {
	url := fmt.Sprintf("%s/objects/states/%s?limit=1", o.driverhub_host, objectId)
	var paginated PaginatedStateRecord
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		Get(url)
	if err != nil {
		return state, err
	}

	if resp.StatusCode() >= 400 {
		return state, errors.New(resp.String())
	}

	err = json.Unmarshal(resp.Body(), &paginated)
	if err != nil {
		return state, err
	}

	if len(paginated.Items) > 0 {
		state = paginated.Items[0]
	}

	return state, nil
}

// UpdateResultAttributes implements ObjectController.
func (o *objectController) UpdateResultAttributes(executionID string, attributes map[string]string) error {
	url := fmt.Sprintf("%s/objects/actions/executions/%s", o.driverhub_host, executionID)
	body := map[string]map[string]string{"result": attributes}
	_, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(body).
		Put(url)
	return err
}

// Increment implements ObjectController.
func (o *objectController) Increment(objectId string) error {
	url := fmt.Sprintf("%s/objects/states/%s/increment", o.driverhub_host, objectId)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		Put(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.String())
	}
	return nil
}

// Decrement implements ObjectController.
func (o *objectController) Decrement(objectId string) error {
	url := fmt.Sprintf("%s/objects/states/%s/decrement", o.driverhub_host, objectId)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		Put(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.String())
	}
	return nil
}

type EventTypeResponse struct {
	Domain             string `json:"domain"`
	DisplayName        string `json:"display_name"`
	DisplayDescription string `json:"display_description"`
	EventType          string `json:"event_type"`
	IsHidden           bool   `json:"is_hidden"`
	EventLevel         string `json:"event_level"`
	Color              string `json:"color"`
	ShowColor          bool   `json:"show_color"`
	Origin             string `json:"origin"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type EventTypesBatchResponse struct {
	Successful []EventTypeResponse `json:"successful"`
	Failed     []EventTypeResponse `json:"failed"`
}

func (o *objectController) AddEventTypes(eventTypes []EventType) error {

	url := fmt.Sprintf("%s/objects/events/types/batch", o.driverhub_host)

	if len(eventTypes) == 0 {
		return errors.New("event types cannot be empty")
	}

	batchSize := 30
	numEventTypes := len(eventTypes)

	for i := 0; i < numEventTypes; i += batchSize {
		end := i + batchSize
		if end > numEventTypes {
			end = numEventTypes
		}
		batch := eventTypes[i:end]

		resp, err := o.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Auth-Token", o.token).
			SetBody(batch).
			Post(url)

		if err != nil {
			return err
		}

		if resp.StatusCode() == 404 {
			// If the endpoint is not found, we assume that the driverhub does not support batch event type creation.
			// We will fall back to the old method of creating event types one by one.
			return o.AddEventTypesFallback(batch)
		}

		if resp.StatusCode() == 201 || resp.StatusCode() == 207 {
			// If the response is 201 or 207, we assume that the event types were created successfully.
			// We will return nil to indicate success.
			var eventTypesBatch EventTypesBatchResponse
			err = json.Unmarshal(resp.Body(), &eventTypesBatch)
			if err != nil {
				return err
			}
			for _, failed := range eventTypesBatch.Failed {
				logger.Logger().Error(fmt.Sprintf("failed to post event type: %s/%s", failed.Domain, failed.EventType))
			}
		}

		if resp.StatusCode() >= 400 {
			content := resp.String()
			if strings.Contains(content, "all event types failed to create") {
				return errors.New("all event types failed to create")
			}
			return errors.New(content)
		}

	}
	return nil
}

// AddEventTypes implements ObjectController.
func (o *objectController) AddEventTypesFallback(eventTypes []EventType) error {

	if len(eventTypes) == 0 {
		return errors.New("event types cannot be empty")
	}

	for _, e := range eventTypes {
		if e.EventType == "" {
			return errors.New("event type cannot be empty")
		}
		if e.Domain == "" {
			return errors.New("domain cannot be empty")
		}

		url := fmt.Sprintf("%s/objects/events/types/%s/%s", o.driverhub_host, e.Domain, e.EventType)
		resp, err := o.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Auth-Token", o.token).
			SetBody(e).
			Post(url)
		if err != nil {
			return err
		}
		if resp.StatusCode() >= 400 {
			content := resp.String()
			if strings.Contains(content, "Duplicate entry") {
				continue
			}
			return errors.New(content)
		}
	}
	return nil
}

// DisabledObject implements ObjectController.
func (o *objectController) DisabledObject(objectId string) error {
	url := fmt.Sprintf("%s/objects/%s/disabled", o.driverhub_host, objectId)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		Put(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.String())
	}
	return nil
}
func (o *objectController) EnabledObject(objectId string) error {
	url := fmt.Sprintf("%s/objects/%s/enabled", o.driverhub_host, objectId)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		Put(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return errors.New(resp.String())
	}
	return nil
}

// GetDriverKey implements ObjectController.
func (o *objectController) GetDriverKey() string {
	return o.driver_key
}

// GetDriverhubHost implements ObjectController.
func (o *objectController) GetDriverhubHost() string {
	return o.driverhub_host
}

// UpdateStateAttributes implements ObjectController.
func (o *objectController) UpdateStateAttributes(objectId string, attributes map[string]string) error {
	url := fmt.Sprintf("%s/objects/states/%s", o.driverhub_host, objectId)
	body := map[string]map[string]string{"state_additional_properties": attributes}
	_, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(body).
		Put(url)
	return err
}

type UpdateStateAttributesBatchRequest struct {
	Changes []ObjectStateChange `json:"changes"`
}

func (o *objectController) UpdateStateAttributesBatch(objectsStates []ObjectStateChange) error {
	url := fmt.Sprintf("%s/objects/states_batch", o.driverhub_host)
	body := UpdateStateAttributesBatchRequest{
		Changes: objectsStates,
	}
	_, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(body).
		Put(url)
	return err
}

// NewAction implements ObjectController.
func (o *objectController) NewAction(action ObjectAction) error {
	url := fmt.Sprintf("%s/objects/actions", o.driverhub_host)

	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(action).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return errors.New(resp.String())
	}

	return nil
}

type wsMessage struct {
	EventType string `json:"event_type"`
	Data      any    `json:"data"`
	Domain    string `json:"domain"`
}

// ListenActionRequests implements ObjectController.
func (o *objectController) ListenActionRequests() error {
	_logger := logger.Logger()

	// Convert the URL using the utility function
	url := tools.ConvertToWebSocketURL(o.driverhub_host, "objects/ws")

	// Create a custom dialer that accepts self-signed certificates
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c, _, err := dialer.Dial(url, http.Header{
		"X-Auth-Token": []string{o.token},
	})
	if err != nil {
		return err
	}
	defer c.Close()

	done := make(chan struct{})

	eventbus.Pubsub.Subscribe("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", func(data interface{}) {
		domain := reflect.ValueOf(data).FieldByName("Domain")
		if domain.IsValid() {
			err = c.WriteJSON(wsMessage{EventType: "REQUEST_SUBSCRIPTION_TO_DOMAIN", Domain: domain.String()})
			if err != nil {
				_logger.Error(err)
			}
		}
	})

	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return nil
		}
		msg := wsMessage{}
		json.Unmarshal(message, &msg)
		if msg.EventType == "REQUEST_ACTION_EXECUTION" {
			go eventbus.Pubsub.Publish("REQUEST_ACTION_EXECUTION", msg.Data)
		}
	}

}

type newObjectRequest struct {
	ID               string   `json:"id"`
	Domain           string   `json:"domain"`
	Name             string   `json:"name"`
	Tags             []string `json:"tags"`
	Type             string   `json:"type"`
	DeviceID         int      `json:"device_id"`
	Enabled          bool     `json:"enabled"`
	StatesAvailable  []string `json:"states_available"`
	EventsAvailable  []string `json:"events_available"`
	ActionsAvailable []string `json:"actions_available"`
}

// CreateObject implements ObjectController.
func (o *objectController) CreateObject(obj RegistrableObject) error {
	req := newObjectRequest{}

	req.ID = obj.GetMetadata().ObjectID
	req.Domain = obj.GetMetadata().Domain
	req.Name = obj.GetMetadata().Name
	req.Tags = obj.GetMetadata().Tags
	req.Type = obj.GetMetadata().Type
	req.Enabled = true
	req.DeviceID, _ = strconv.Atoi(obj.GetMetadata().DeviceID)
	req.EventsAvailable = []string{}
	req.StatesAvailable = []string{}
	req.ActionsAvailable = []string{}

	req.StatesAvailable = append(req.StatesAvailable, obj.GetAvailableStates()...)

	for _, action := range obj.GetAvailableActions() {
		req.ActionsAvailable = append(req.ActionsAvailable, action.Action)
	}

	url := fmt.Sprintf("%s/objects", o.driverhub_host)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(req).
		Post(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("error creating object: %s", resp.String())
	}

	return nil

}

// SetState implements ObjectController.
func (o *objectController) SetState(objectId, state string) error {
	url := fmt.Sprintf("%s/objects/states/%s", o.driverhub_host, objectId)
	body := map[string]string{"state": state}
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(body).
		Put(url)

	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		if strings.Contains(resp.String(), "object is disabled") {
			return nil
		}
		return fmt.Errorf("error setting state: %s", resp.String())
	}
	return nil
}

func NewObjectController(driverhubHost string, driverKey string) ObjectController {
	if driverhubHost == "" {
		panic("driverhub host cannot be empty")
	}

	fileData, _ := tools.GetDriverNetsocsDotJsonContent("driver.netsocs.json")

	if !strings.HasPrefix(driverhubHost, "http") && !strings.HasPrefix(driverhubHost, "https") {
		driverhubHost = fmt.Sprintf("http://%s", driverhubHost)
	}

	return &objectController{
		driverhub_host: driverhubHost,
		driver_key:     driverKey,
		httpClient:     resty.New(),
		token:          fileData.Token,
	}
}
