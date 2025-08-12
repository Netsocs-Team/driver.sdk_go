package client

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type DeviceState string

const (
	DeviceStateOnline                DeviceState = "Online"
	DeviceStateOffline               DeviceState = "Offline"
	DeviceStateConfigurationFailure  DeviceState = "ConfigurationFailure"
	DeviceStateAuthenticationFailure DeviceState = "AuthenticationFailure"
	DeviceStateDuplicatedDevice      DeviceState = "DuplicatedDevice"
	DeviceStateUnknown               DeviceState = "Unknown"
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

type DeviceStateItem struct {
	Id       string      `json:"id"`
	DeviceID int         `json:"device_id"`
	State    DeviceState `json:"state"`
	Datetime string      `json:"datetime"`
}

type DeviceStateResponse struct {
	Items    []DeviceStateItem `json:"items"`
	Metadata Metadata          `json:"_metadata"`
}

type Metadata struct {
	Total  int `json:"total_items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (d *NetsocsDriverClient) GetDeviceState(deviceId int) (DeviceStateItem, error) {
	resp, err := resty.New().R().SetHeader("X-Auth-Token", d.token).Get(d.driverHubHost + "/devices/" + "states/" + strconv.Itoa(deviceId))
	if err != nil {
		return DeviceStateItem{}, err
	}

	var deviceState DeviceStateResponse
	if err := json.Unmarshal(resp.Body(), &deviceState); err != nil {
		return DeviceStateItem{}, err
	}

	if len(deviceState.Items) == 0 {
		return DeviceStateItem{}, errors.New("no device state found")
	}

	return deviceState.Items[0], nil
}

func (d *NetsocsDriverClient) SetDeviceState(deviceId int, state DeviceState) (DeviceStateResponse, error) {
	resp, err := resty.New().R().SetHeader("X-Auth-Token", d.token).SetBody(map[string]interface{}{
		"state": state,
	}).Put(d.driverHubHost + "/devices/" + "states/" + strconv.Itoa(deviceId))
	if err != nil {
		return DeviceStateResponse{}, err
	}

	if resp.IsError() {
		return DeviceStateResponse{}, errors.New(resp.String())
	}

	return DeviceStateResponse{}, nil
}
