package objects

import "errors"

var ErrMethodNotImplemented = errors.New("method not implemented")
var ErrAttributeNotFound = errors.New("attribute not found")
var ErrObjectNotFound = errors.New("object not found")
var ErrObjectAlreadyExists = errors.New("object already exists")

var ErrDomainMandatory = errors.New("domain is mandatory")
var ErrObjectIdMandatory = errors.New("object_id is mandatory")
var ErrNameMandatory = errors.New("name is mandatory")
var ErrActionsMandatory = errors.New("actions are mandatory")
var ErrDeviceIdMandatory = errors.New("device_id is mandatory")
