package device

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type StateRecord struct {
	ID        string `json:"id,omitempty"`
	DeviceID  string `json:"device_id"`
	State     string `json:"state,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type PaginatedStateRecord struct {
	Items []StateRecord `json:"items"`
}

type Device interface {
	GetID() string
	GetName() string
	SetState(state string) error
	GetState() (StateRecord, error)
}

type device struct {
	driverhub_host string
	driver_key     string
	token          string
	httpClient     *resty.Client
	deviceId       string
}

func (o *device) SetDeviceID(deviceId string) {
	o.deviceId = deviceId
}

func (o *device) GetState() (state StateRecord, err error) {
	url := fmt.Sprintf("%s/devices/states/%s?limit=1", o.driverhub_host, o.deviceId)
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

func (o *device) SetState(state string) error {
	url := fmt.Sprintf("%s/devices/states/%s", o.driverhub_host, o.deviceId)
	body := map[string]string{"state": state}
	_, err := o.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", o.token).
		SetBody(body).
		Put(url)
	return err
}
