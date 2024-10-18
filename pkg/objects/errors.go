package objects

import "errors"

var ErrMethodNotImplemented = errors.New("method not implemented")
var ErrAttributeNotFound = errors.New("attribute not found")
var ErrObjectNotFound = errors.New("object not found")
var ErrObjectAlreadyExists = errors.New("object already exists")
