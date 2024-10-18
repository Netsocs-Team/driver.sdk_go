package objects

import "fmt"

const SWITCH_OBJECT_TYPE = "switch"

type switchObject struct {
	Object
	// Boolean value that represents the state of the switch
	// true if the switch is on, false if the switch is off
	IsOn    bool `json:"isOn"`
	TurnOn  func(changeStateToOn func()) error
	TurnOff func(changeStateToOff func()) error
}

// GetName implements ObjectRunner.
func (s *switchObject) GetName() string {
	return s.Name
}

// GetType implements ObjectRunner.
func (s *switchObject) GetType() string {
	return s.Type
}

// GetAvailableMethods implements ObjectRunner.
func (s *switchObject) GetAvailableMethods() []string {
	return []string{SWITCH_TURN_ON, SWITCH_TURN_OFF}
}

// GetID implements ObjectRunner.
func (s *switchObject) GetID() string {
	return s.ID
}

// GetAttribute implements ObjectRunner.
func (s *switchObject) GetAttribute(attributeName string) (interface{}, error) {
	switch attributeName {
	case "isOn":
		return s.IsOn, nil
	}
	return nil, ErrAttributeNotFound
}

// RunMethod implements ObjectRunner.
func (s *switchObject) RunMethod(methodName string, value string) (interface{}, error) {
	switch methodName {
	case SWITCH_TURN_ON:
		err := s.turnOn()
		if err != nil {
			return nil, err
		}
		return nil, nil
	case SWITCH_TURN_OFF:
		err := s.turnOff()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, ErrMethodNotImplemented
}

func (s *switchObject) turnOn() error {
	return s.TurnOn(func() {
		s.Object.CommunicationToHandlerChannel <- commandToObjectHandler{
			ObjectID: s.ID,
			Command:  _CHANGE_ATTRIBUTE_COMMAND,
			Params:   `{"key":"isOn","value":"true"}`,
		}
	})
}

func (s *switchObject) turnOff() error {
	return s.TurnOff(func() {
		s.Object.CommunicationToHandlerChannel <- commandToObjectHandler{
			ObjectID: s.ID,
			Command:  _CHANGE_ATTRIBUTE_COMMAND,
			Params:   `{"key":"isOn","value":"false"}`,
		}
	})
}

// SetAttribute implements ObjectRunner.
func (s *switchObject) SetAttribute(attributeName string, value string) error {
	switch attributeName {
	case "isOn":
		if value == "true" {
			s.IsOn = true
		} else {
			s.IsOn = false
		}
		return nil
	}
	return ErrAttributeNotFound
}

func (s *switchObject) GetIcon() string {
	return s.Icon
}

func (s *switchObject) GetDeviceID() int {
	return s.DeviceID
}

func (s *switchObject) GetCommandsChannel() chan commandToObjectHandler {
	return s.CommunicationToHandlerChannel
}

func NewSwitchObject(name string, id string, deviceId int, icon string) *switchObject {
	swch := switchObject{}
	swch.Object.CommunicationToHandlerChannel = make(chan commandToObjectHandler)
	swch.ID = fmt.Sprintf("%d:%s", deviceId, id)
	swch.Name = name
	swch.Type = SWITCH_OBJECT_TYPE
	swch.DeviceID = deviceId
	swch.Icon = icon
	return &swch
}
