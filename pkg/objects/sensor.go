package objects

import (
	"errors"
	"fmt"
)

type sensorObject struct {
	//sensorType        SensorObjectType
	setup    SetupFunction
	metatada ObjectMetadata
	//unitOfMeasurement string
	controller            ObjectController
	eventTypes            []EventType
	alarmDetectorBypass   func(this SensorObject, controller ObjectController, payload AlarmDetectorBypassPayload) (map[string]string, error)
	alarmDetectorUnbypass func(this SensorObject, controller ObjectController, payload AlarmDetectorUnbypassPayload) (map[string]string, error)
	customAction          func(this SensorObject, controller ObjectController, payload CustomActionPayload) (map[string]string, error)
}

// Decrement implements SensorObject.
func (s *sensorObject) Decrement() error {
	return s.controller.Decrement(s.metatada.ObjectID)
}

// Increment implements SensorObject.
func (s *sensorObject) Increment() error {
	return s.controller.Increment(s.metatada.ObjectID)
}

// SetSensorType implements SensorObject.
func (s *sensorObject) SetSensorType(sensorType SensorObjectType) error {
	return s.UpdateStateAttributes(map[string]string{"sensor_type": string(sensorType)})
}

// SetUnitOfMeasurement implements SensorObject.
func (s *sensorObject) SetUnitOfMeasurement(unitOfMeasurement string) error {
	return s.UpdateStateAttributes(map[string]string{"unit_of_measurement": unitOfMeasurement})
}

// UpdateStateAttributes implements SensorObject.
func (s *sensorObject) UpdateStateAttributes(attributes map[string]string) error {
	return s.controller.UpdateStateAttributes(s.metatada.ObjectID, attributes)
}

// SetState implements SensorObject.
func (s *sensorObject) SetState(state string) error {
	if s.controller == nil {
		return errors.New("controller not set")
	}
	return s.controller.SetState(s.metatada.ObjectID, state)
}

// AddEventTypes implements SensorObject.
func (s *sensorObject) AddEventTypes(eventTypes []EventType) error {
	if s.controller == nil {
		s.eventTypes = eventTypes
		return nil
	}
	for i := range eventTypes {
		e := eventTypes[i]
		e.Domain = s.metatada.Domain
		e.Origin = "driver"
		eventTypes[i] = e

	}
	return s.controller.AddEventTypes(eventTypes)
}

// SetValue implements SensorObject.
func (s *sensorObject) SetValue(value string) error {
	if s.controller != nil {
		return s.controller.UpdateStateAttributes(s.metatada.ObjectID, map[string]string{"value": value})
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

const SENSOR_ACTION_BYPASS = "bypass"
const SENSOR_ACTION_UNBYPASS = "unbypass"
const SENSOR_CUSTOM_ACTION = "custom"

type SensorObject interface {
	RegistrableObject
	SetValue(value string) error
	SetSensorType(sensorType SensorObjectType) error
	SetUnitOfMeasurement(unitOfMeasurement string) error
	Increment() error
	Decrement() error
}

// GetAvailableActions implements RegistrableObject.
func (s *sensorObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: SENSOR_ACTION_BYPASS,
			Domain: s.metatada.Domain,
		},
		{
			Action: SENSOR_ACTION_UNBYPASS,
			Domain: s.metatada.Domain,
		},
		{
			Action: SENSOR_CUSTOM_ACTION,
			Domain: s.metatada.Domain,
		},
	}
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
func (s *sensorObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
	switch action {
	case SENSOR_ACTION_BYPASS:
		if s.alarmDetectorBypass == nil {
			return nil, fmt.Errorf("alarm detector bypass not set")
		}
		return s.alarmDetectorBypass(s, s.controller, AlarmDetectorBypassPayload{})
	case SENSOR_ACTION_UNBYPASS:
		if s.alarmDetectorUnbypass == nil {
			return nil, fmt.Errorf("alarm detector unbypass not set")
		}
		return s.alarmDetectorUnbypass(s, s.controller, AlarmDetectorUnbypassPayload{})
	case SENSOR_CUSTOM_ACTION:
		if s.customAction == nil {
			return nil, fmt.Errorf("custom action not set")
		}
		return s.customAction(s, s.controller, CustomActionPayload{})
	}
	return nil, fmt.Errorf("action %s not found", action)
}

// Setup implements RegistrableObject.
func (s *sensorObject) Setup(oc ObjectController) error {
	s.controller = oc
	if s.setup == nil {
		return nil
	}
	return s.setup(s, oc)
}

type AlarmDetectorBypassPayload struct {
}

type AlarmDetectorUnbypassPayload struct {
}

type CustomActionPayload struct {
	Action  string `json:"action"`
	Payload string `json:"payload"`
}

type NewSensorObjectParams struct {
	Metadata ObjectMetadata
	SetupFn  SetupFunction

	AlarmDetectorBypass   func(this SensorObject, controller ObjectController, payload AlarmDetectorBypassPayload) (map[string]string, error)
	AlarmDetectorUnbypass func(this SensorObject, controller ObjectController, payload AlarmDetectorUnbypassPayload) (map[string]string, error)
	CustomAction          func(this SensorObject, controller ObjectController, payload CustomActionPayload) (map[string]string, error)
}

func NewSensorObject(params NewSensorObjectParams) SensorObject {
	return &sensorObject{
		metatada:              params.Metadata,
		setup:                 params.SetupFn,
		alarmDetectorBypass:   params.AlarmDetectorBypass,
		alarmDetectorUnbypass: params.AlarmDetectorUnbypass,
		customAction:          params.CustomAction,
	}
}
