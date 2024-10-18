package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Spy struct {
	mock.Mock
}

func (s *Spy) Test(func()) error {
	args := s.Called()
	return args.Error(0)
}

func TestRunTurnOn(t *testing.T) {
	s := new(Spy)
	s.On("Test").Return(nil)
	_switchObject := switchObject{
		TurnOn: s.Test,
	}

	_, err := _switchObject.RunMethod(SWITCH_TURN_ON, "")

	assert.Nil(t, err)
	s.AssertExpectations(t)
}

func TestRunTurnOnAndExpectErr(t *testing.T) {
	s := new(Spy)
	s.On("Test").Return(assert.AnError)
	_switchObject := switchObject{
		TurnOn: s.Test,
	}

	_, err := _switchObject.RunMethod(SWITCH_TURN_ON, "")

	assert.Equal(t, assert.AnError, err)
	s.AssertExpectations(t)
}

func TestRunTurnOff(t *testing.T) {
	s := new(Spy)
	s.On("Test").Return(nil)
	_switchObject := switchObject{
		TurnOff: s.Test,
	}

	_, err := _switchObject.RunMethod(SWITCH_TURN_OFF, "")

	assert.Nil(t, err)
	s.AssertExpectations(t)
}

func TestRunTurnOffAndExpectErr(t *testing.T) {
	s := new(Spy)
	s.On("Test").Return(assert.AnError)
	_switchObject := switchObject{
		TurnOff: s.Test,
	}

	_, err := _switchObject.RunMethod(SWITCH_TURN_OFF, "")

	assert.Equal(t, assert.AnError, err)
	s.AssertExpectations(t)
}

func TestRunNotImplementMethod(t *testing.T) {
	_switchObject := switchObject{}

	_, err := _switchObject.RunMethod("not_implemented", "")

	assert.Equal(t, ErrMethodNotImplemented, err)
}

func TestGetIsOnAttribute(t *testing.T) {
	_switchObject := switchObject{
		IsOn: true,
	}

	attr, err := _switchObject.GetAttribute("isOn")

	assert.Nil(t, err)
	assert.Equal(t, true, attr)
}

func TestExpectAttributeNotFound(t *testing.T) {
	_switchObject := switchObject{}

	_, err := _switchObject.GetAttribute("not_found")

	assert.Equal(t, ErrAttributeNotFound, err)
}

func TestSetIsOnAttribute(t *testing.T) {
	_switchObject := switchObject{}

	err := _switchObject.SetAttribute("isOn", "true")

	assert.Nil(t, err)
	assert.Equal(t, true, _switchObject.IsOn)

	err = _switchObject.SetAttribute("isOn", "false")
	assert.Nil(t, err)
	assert.Equal(t, false, _switchObject.IsOn)
}

func TestExpectAttributeNotFoundWhenSetAttribute(t *testing.T) {
	_switchObject := switchObject{}

	err := _switchObject.SetAttribute("not_found", "true")

	assert.Equal(t, ErrAttributeNotFound, err)
}

func TestGetID(t *testing.T) {
	_switchObject := switchObject{
		Object: Object{
			ID: "netsocs_hardware:123456:relay_1",
		},
	}

	assert.Equal(t, "netsocs_hardware:123456:relay_1", _switchObject.GetID())
}
