package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type objectMock struct {
	mock.Mock
}

// GetCommandsChannel implements ObjectRunner.
func (o *objectMock) GetCommandsChannel() chan commandToObjectHandler {
	panic("unimplemented")
}

// GetDeviceID implements ObjectRunner.
func (o *objectMock) GetDeviceID() int {
	panic("unimplemented")
}

// GetIcon implements ObjectRunner.
func (o *objectMock) GetIcon() string {
	panic("unimplemented")
}

// GetName implements ObjectRunner.
func (o *objectMock) GetName() string {
	panic("unimplemented")
}

// GetType implements ObjectRunner.
func (o *objectMock) GetType() string {
	panic("unimplemented")
}

// GetAttribute implements ObjectRunner.
func (o *objectMock) GetAttribute(attributeName string) (interface{}, error) {
	panic("unimplemented")
}

// GetAvailableMethods implements ObjectRunner.
func (o *objectMock) GetAvailableMethods() []string {
	args := o.Called()
	return args.Get(0).([]string)
}

// GetID implements ObjectRunner.
func (o *objectMock) GetID() string {
	args := o.Called()
	return args.String(0)
}

// RunMethod implements ObjectRunner.
func (o *objectMock) RunMethod(methodName string, value string) (interface{}, error) {
	panic("unimplemented")
}

// SetAttribute implements ObjectRunner.
func (o *objectMock) SetAttribute(attributeName string, value string) error {
	panic("unimplemented")
}

func TestIsConfigForObject(t *testing.T) {
	o := &objectHandler{}
	objMock := new(objectMock)

	objMock.On("GetID").Return("test")
	objMock.On("GetAvailableMethods").Return([]string{"method1", "method2"})

	o.AppendObject(objMock)
	assert.True(t, o.IsConfigForObject("method1"))
	assert.True(t, o.IsConfigForObject("method2"))
	assert.False(t, o.IsConfigForObject("method3"))

}
