package objects

import (
	"encoding/json"
)

type ObjectRunner interface {
	RunMethod(methodName string, value string) (interface{}, error)
	SetAttribute(attributeName string, value string) error
	GetAttribute(attributeName string) (interface{}, error)
	GetAvailableMethods() []string
	GetName() string
	GetType() string
	GetID() string
	GetIcon() string
	GetDeviceID() int
	GetCommandsChannel() chan commandToObjectHandler
}

type ObjectHandler interface {
	AppendObject(obj ObjectRunner) error
	BuildID(descriptiveName string) string
	IsConfigForObject(methodName string) bool
	CallMethod(methodName string, value string) (interface{}, error)
}

type objectHandler struct {
	objects map[string]ObjectRunner
}

type callMethodValue struct {
	ObjectID string `json:"objectID"`
}

// CallMethod implements ObjectHandler.
func (o *objectHandler) CallMethod(methodName string, value string) (interface{}, error) {
	data := &callMethodValue{}
	err := json.Unmarshal([]byte(value), data)
	if err != nil {
		return nil, err
	}
	obj, ok := o.objects[data.ObjectID]
	if !ok {
		return nil, ErrObjectNotFound
	}

	isNativeMethod := false
	for _, method := range NATIVE_OBJECT_CONFIG_KEYS {
		if method == methodName {
			isNativeMethod = true
			break
		}
	}

	if !isNativeMethod {
		return obj.RunMethod(methodName, value)
	}

	switch methodName {
	case SET_ATTRIBUTE_TO_OBJECT:
		req := &changeAttributeCommandParams{}
		err := json.Unmarshal([]byte(value), req)
		if err != nil {
			return nil, err
		}
		err = obj.SetAttribute(req.Key, req.Value)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// IsConfigForObject implements ObjectHandler.
func (o *objectHandler) IsConfigForObject(methodName string) bool {
	methodsAvailable := NATIVE_OBJECT_CONFIG_KEYS
	for _, obj := range o.objects {
		methodsAvailable = append(methodsAvailable, obj.GetAvailableMethods()...)
	}
	for _, method := range methodsAvailable {
		if method == methodName {
			return true
		}
	}
	return false
}

// BuildID implements ObjectHandler.
func (o *objectHandler) BuildID(descriptiveName string) string {
	panic("unimplemented")
}

// AppendObject implements ObjectHandler.
func (o *objectHandler) AppendObject(obj ObjectRunner) error {
	if o.objects == nil {
		o.objects = make(map[string]ObjectRunner)
	}

	if _, ok := o.objects[obj.GetID()]; ok {
		return ErrObjectAlreadyExists
	}

	o.objects[obj.GetID()] = obj
	return nil
}

var globalObjectHandler ObjectHandler

func GetObjectHandler() ObjectHandler {
	if globalObjectHandler == nil {
		globalObjectHandler = &objectHandler{}
	}
	return globalObjectHandler
}
