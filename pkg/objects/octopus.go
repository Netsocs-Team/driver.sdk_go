package objects

import (
	"errors"
	"fmt"
	"strconv"
)

type octopusObject struct {
	setup      SetupFunction
	metatada   ObjectMetadata
	controller ObjectController
	eventTypes []EventType
	relayOnFn  func(this OctopusObject, controller ObjectController, payload RelayOnPayload) (map[string]string, error)
	relayOffFn func(this OctopusObject, controller ObjectController, payload RelayOffPayload) (map[string]string, error)
}

const (
	OCTOPUS_ACTION_RELAY_ON  = "octopus.action.turn_on"
	OCTOPUS_ACTION_RELAY_OFF = "octopus.action.turn_off"
)

// UpdateStateAttributes implements SensorObject.
func (s *octopusObject) UpdateStateAttributes(attributes map[string]string) error {
	return s.controller.UpdateStateAttributes(s.metatada.ObjectID, attributes)
}

// SetState implements SensorObject.
func (s *octopusObject) SetState(state string) error {
	if s.controller == nil {
		return errors.New("controller not set")
	}
	return s.controller.SetState(s.metatada.ObjectID, state)
}

// AddEventTypes implements SensorObject.
func (s *octopusObject) AddEventTypes(eventTypes []EventType) error {
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
func (s *octopusObject) SetOffline(status bool) error {
	if s.controller != nil {
		return s.controller.UpdateStateAttributes(s.metatada.ObjectID, map[string]string{"offline": strconv.FormatBool(status)})
	}
	return errors.New("controller not set")
}

type OctopusObject interface {
	RegistrableObject
	SetOffline(status bool) error
}

// GetAvailableActions implements RegistrableObject.
func (s *octopusObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: OCTOPUS_ACTION_RELAY_ON,
			Domain: s.metatada.Domain,
		},
		{
			Action: OCTOPUS_ACTION_RELAY_OFF,
			Domain: s.metatada.Domain,
		},
	}
}

// GetAvailableStates implements RegistrableObject.
func (s *octopusObject) GetAvailableStates() []string {
	return []string{}
}

// GetMetadata implements RegistrableObject.
func (s *octopusObject) GetMetadata() ObjectMetadata {
	s.metatada.Type = "octopus"
	return s.metatada
}

// RunAction implements RegistrableObject.
func (s *octopusObject) RunAction(id, action string, payload []byte) (map[string]string, error) {
	switch action {
	case OCTOPUS_ACTION_RELAY_ON:
		return s.relayOnFn(s, s.controller, RelayOnPayload{RelayID: string(payload)})
	case OCTOPUS_ACTION_RELAY_OFF:
		return s.relayOffFn(s, s.controller, RelayOffPayload{RelayID: string(payload)})
	}
	return nil, fmt.Errorf("action %s not found", action)
}

// Setup implements RegistrableObject.
func (s *octopusObject) Setup(oc ObjectController) error {
	s.controller = oc
	if s.setup == nil {
		return nil
	}
	return s.setup(s, oc)
}

type RelayOnPayload struct {
	RelayID string `json:"relay_id"`
}

type RelayOffPayload struct {
	RelayID string `json:"relay_id"`
}

type NewOctopusObjectParams struct {
	Metadata   ObjectMetadata
	SetupFn    SetupFunction
	RelayOnFn  func(this OctopusObject, controller ObjectController, payload RelayOnPayload) (map[string]string, error)
	RelayOffFn func(this OctopusObject, controller ObjectController, payload RelayOffPayload) (map[string]string, error)
}

func NewOctopusObject(params NewOctopusObjectParams) OctopusObject {
	return &octopusObject{
		metatada:   params.Metadata,
		setup:      params.SetupFn,
		relayOnFn:  params.RelayOnFn,
		relayOffFn: params.RelayOffFn,
	}
}
