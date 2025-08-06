package client

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Device struct {
	Username        string                 `json:"username"`
	Password        string                 `json:"password"`
	IpAddressPublic string                 `json:"ipAddressPublic"`
	Port            int                    `json:"port"`
	ID              string                 `json:"id"`
	IDModel         int                    `json:"idModel"`
	ChildID         string                 `json:"childId"`
	IDBrand         int                    `json:"idBrand"`
	IDManufacturer  int                    `json:"idManufacturer"`
	IDDeviceGroup   int                    `json:"idDeviceGroup"`
	IDSubSystem     int                    `json:"idSubSystem"`
	Params          map[string]interface{} `json:"params"`
}

type DeviceState struct {
	Id            int    `json:"id"`
	DeviceID      int    `json:"device_id"`
	State         string `json:"state"`
	Datetime      string `json:"datetime"`
	PreviousState string `json:"prev_state"`
}

func (d *NetsocsDriverClient) GetDeviceState(deviceId int) (DeviceState, error) {
	resp, err := resty.New().R().SetHeader("X-Auth-Token", d.token).Get(d.driverHubHost + "/devices/" + strconv.Itoa(deviceId) + "/state")
	if err != nil {
		return DeviceState{}, err
	}

	var deviceState []DeviceState
	if err := json.Unmarshal(resp.Body(), &deviceState); err != nil {
		return DeviceState{}, err
	}

	if len(deviceState) == 0 {
		return DeviceState{}, errors.New("no device state found")
	}

	return deviceState[0], nil
}

func (d *NetsocsDriverClient) SetDeviceState(deviceId int, state string) (DeviceState, error) {
	resp, err := resty.New().R().SetHeader("X-Auth-Token", d.token).SetBody(map[string]interface{}{
		"state": state,
	}).Post(d.driverHubHost + "/devices/" + strconv.Itoa(deviceId) + "/state")
	if err != nil {
		return DeviceState{}, err
	}

	if resp.IsError() {
		return DeviceState{}, errors.New(resp.String())
	}

	return DeviceState{}, nil
}
