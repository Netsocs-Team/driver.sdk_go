package objects

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type objectController struct {
	driverhub_host string
	driver_key     string
	httpClient     *resty.Client
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
func (o *objectController) UpdateStateAttributes(objectId string, attributes map[string]interface{}) error {
	url := fmt.Sprintf("%s/objects/states/%s", o.driverhub_host, objectId)
	body := map[string]map[string]interface{}{"state_additional_properties": attributes}
	_, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Put(url)
	return err
}

// NewAction implements ObjectController.
func (o *objectController) NewAction(action ObjectAction) error {
	url := fmt.Sprintf("%s/objects/actions", o.driverhub_host)

	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
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
}

// ListenActionRequests implements ObjectController.
func (o *objectController) ListenActionRequests() error {
	url := strings.ReplaceAll(o.driverhub_host, "https", "wss")
	url = strings.ReplaceAll(url, "http", "ws")

	url = fmt.Sprintf("%s/objects/ws", url)

	c, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return err
	}
	defer c.Close()

	done := make(chan struct{})

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
	ID       string   `json:"id"`
	Domain   string   `json:"domain"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	Type     string   `json:"type"`
	DeviceID int      `json:"device_id"`
	Enabled  bool     `json:"enabled"`
}

// CreateObject implements ObjectController.
func (o *objectController) CreateObject(obj RegistrableObject) error {
	req := newObjectRequest{}

	req.ID = obj.GetMetadata().ObjectID
	req.Domain = obj.GetMetadata().Domain
	req.Name = obj.GetMetadata().Name
	req.Tags = obj.GetMetadata().Tags
	req.Type = obj.GetMetadata().Type
	req.DeviceID, _ = strconv.Atoi(obj.GetMetadata().DeviceID)

	url := fmt.Sprintf("%s/objects", o.driverhub_host)
	resp, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
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

	return &objectController{
		driverhub_host: driverhubHost,
		driver_key:     driverKey,
		httpClient:     resty.New(),
	}
}
