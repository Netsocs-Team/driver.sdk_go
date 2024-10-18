package objects

import (
	"fmt"
	"strconv"
)

const GPSTrackerType = "gps_tracker"

type GPSTracker struct {
	Object
	BatteryLevel int     `json:"batteryLevel"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	// in meters
	LocationAccuracy int    `json:"locationAccuracy"`
	LocationName     string `json:"locationName"`
}

func (g *GPSTracker) SetBatteryLevel(batteryLevel int) {
	g.CommunicationToHandlerChannel <- commandToObjectHandler{
		Command: _CHANGE_ATTRIBUTE_COMMAND,
		Params:  `{"key":"batteryLevel","value":"` + fmt.Sprintf("%d", batteryLevel) + `"}`,
	}
}

func (g *GPSTracker) SetLatitude(latitude float64) {
	g.CommunicationToHandlerChannel <- commandToObjectHandler{
		Command: _CHANGE_ATTRIBUTE_COMMAND,
		Params:  `{"key":"latitude","value":"` + fmt.Sprintf("%f", latitude) + `"}`,
	}
}

func (g *GPSTracker) SetLongitude(longitude float64) {
	g.CommunicationToHandlerChannel <- commandToObjectHandler{
		Command: _CHANGE_ATTRIBUTE_COMMAND,
		Params:  `{"key":"longitude","value":"` + fmt.Sprintf("%f", longitude) + `"}`,
	}
}

func (g *GPSTracker) SetLocationAccuracy(locationAccuracy int) {
	g.CommunicationToHandlerChannel <- commandToObjectHandler{
		Command: _CHANGE_ATTRIBUTE_COMMAND,
		Params:  `{"key":"locationAccuracy","value":"` + fmt.Sprintf("%d", locationAccuracy) + `"}`,
	}
}

func (g *GPSTracker) SetLocationName(locationName string) {
	g.CommunicationToHandlerChannel <- commandToObjectHandler{
		Command: _CHANGE_ATTRIBUTE_COMMAND,
		Params:  `{"key":"locationName","value":"` + locationName + `"}`,
	}
}

// GetAttribute implements ObjectRunner.
func (g *GPSTracker) GetAttribute(attributeName string) (interface{}, error) {
	switch attributeName {
	case "batteryLevel":
		return g.BatteryLevel, nil
	case "latitude":
		return g.Latitude, nil
	case "longitude":
		return g.Longitude, nil
	case "locationAccuracy":
		return g.LocationAccuracy, nil
	case "locationName":
		return g.LocationName, nil
	}
	return nil, ErrAttributeNotFound
}

// GetAvailableMethods implements ObjectRunner.
func (g *GPSTracker) GetAvailableMethods() []string {
	return []string{}
}

// GetCommandsChannel implements ObjectRunner.
func (g *GPSTracker) GetCommandsChannel() chan commandToObjectHandler {
	return g.CommunicationToHandlerChannel
}

// GetDeviceID implements ObjectRunner.
func (g *GPSTracker) GetDeviceID() int {
	return g.DeviceID
}

// GetID implements ObjectRunner.
func (g *GPSTracker) GetID() string {
	return g.ID
}

// GetIcon implements ObjectRunner.
// Subtle: this method shadows the method (Object).GetIcon of GPSTracker.Object.
func (g *GPSTracker) GetIcon() string {
	return g.Icon
}

// GetName implements ObjectRunner.
func (g *GPSTracker) GetName() string {
	return g.Name
}

// GetType implements ObjectRunner.
func (g *GPSTracker) GetType() string {
	return g.Type
}

// RunMethod implements ObjectRunner.
func (g *GPSTracker) RunMethod(methodName string, value string) (interface{}, error) {
	return nil, ErrMethodNotImplemented
}

// SetAttribute implements ObjectRunner.
func (g *GPSTracker) SetAttribute(attributeName string, value string) error {
	switch attributeName {
	case "batteryLevel":
		v, _ := strconv.Atoi(value)
		g.BatteryLevel = v
	case "latitude":
		v, _ := strconv.ParseFloat(value, 64)
		g.Latitude = v
	case "longitude":
		v, _ := strconv.ParseFloat(value, 64)
		g.Longitude = v
	case "locationAccuracy":
		v, _ := strconv.Atoi(value)
		g.LocationAccuracy = v
	case "locationName":
		g.LocationName = value
	}
	return ErrAttributeNotFound
}

func NewGPSTracker(name, id string, deviceID int, icon string) *GPSTracker {
	return &GPSTracker{
		Object: Object{
			Name:                          name,
			ID:                            fmt.Sprintf("%d:%s", deviceID, id),
			Type:                          GPSTrackerType,
			DeviceID:                      deviceID,
			Icon:                          icon,
			CommunicationToHandlerChannel: make(chan commandToObjectHandler),
		},
	}
}
