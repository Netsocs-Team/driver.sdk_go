package objects

import "errors"

type sensorObject struct {
	sensorType        SensorObjectType
	setup             SetupFunction
	metatada          ObjectMetadata
	unitOfMeasurement string
	controller        ObjectController
}

// SetValue implements SensorObject.
func (s *sensorObject) SetValue(value string) error {
	if s.controller != nil {
		return s.controller.UpdateStateAttributes(s.metatada.ObjectID, map[string]interface{}{"value": value})
	}
	return errors.New("controller not set")
}

type SensorObjectType string

const (
	SensorObjectTypeNumber  SensorObjectType = "" // default
	SensorObjectTypeText    SensorObjectType = "text"
	SensorObjectTypeBinary  SensorObjectType = "binary"
	SensorObjectTypeBattery SensorObjectType = "battery"
)

const SENSOR_STATE_MEASUREMENT = "sensor.state.measurement"
const SENSOR_STATE_TOTAL = "sensor.state.total"
const SENSOR_STATE_TOTAL_INCREASING = "sensor.state.total_increasing"

type SensorObject interface {
	RegistrableObject
	SetValue(value string) error
}

// GetAvailableActions implements RegistrableObject.
func (s *sensorObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{}
}

// GetAvailableStates implements RegistrableObject.
func (s *sensorObject) GetAvailableStates() []string {
	return []string{SENSOR_STATE_MEASUREMENT, SENSOR_STATE_TOTAL, SENSOR_STATE_TOTAL_INCREASING}
}

// GetMetadata implements RegistrableObject.
func (s *sensorObject) GetMetadata() ObjectMetadata {
	s.metatada.Type = "sensor"
	return s.metatada
}

// RunAction implements RegistrableObject.
func (s *sensorObject) RunAction(action string, payload []byte) error {
	return nil
}

// Setup implements RegistrableObject.
func (s *sensorObject) Setup(oc ObjectController) error {
	s.controller = oc
	return s.setup(s, oc)
}

func NewSensorObject(sensorType SensorObjectType, unitOfMeasurement string, objectMetadata ObjectMetadata, setup SetupFunction) SensorObject {
	return &sensorObject{
		metatada: objectMetadata,
		setup:    setup,
	}
}
