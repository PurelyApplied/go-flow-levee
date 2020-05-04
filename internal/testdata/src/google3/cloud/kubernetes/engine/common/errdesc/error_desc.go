// Package errdesc fakes cloud/kubernetes/engine/common/errdesc.
package errdesc

import (
	"fmt"
)

// GKEErrorDescriptor fakes errdesc.GKEErrorDescriptor
type GKEErrorDescriptor struct {
	InternalCode  int
	CanonicalCode int
	DefaultMsg    string
}

var (
	// InternalError  fakes errdesc.InternalError
	InternalError = &GKEErrorDescriptor{
		InternalCode:  99,
		CanonicalCode: 99,
		DefaultMsg:    "InternalError",
	}
	// AlreadyExists fakes errdesc.AlreadyExists
	AlreadyExists = &GKEErrorDescriptor{
		InternalCode:  199,
		CanonicalCode: 299,
		DefaultMsg:    "AlreadyExists",
	}
)

// WithMsg returns a new error of the provided descriptor type, with the
// error message created out of the format string and provided arguments.
// If no formatting arguments provided, the format string is used as the error
// message.
func (ged *GKEErrorDescriptor) WithMsg(format string, a ...interface{}) error {
	err := ged.createErr()
	err.Params = a

	if err.msgStub == "" {
		err.msgStub = format
	} else if format != "" {
		err.msgStub = fmt.Sprintf("%s: %s", err.msgStub, format)
	}

	return err
}

func (ged *GKEErrorDescriptor) createErr() *GKEError {
	return &GKEError{
		InternalCode:  ged.InternalCode,
		CanonicalCode: 100,
		StackTrace:    "fake stacktrace",
		msgStub:       ged.DefaultMsg,
	}
}

// GKEError is an instance of a specific error type.
type GKEError struct {
	InternalCode  int
	CanonicalCode int
	StackTrace    string
	Conditions    []*interface{}
	msgStub       string

	// Do not use these fields directly.
	// Use instead either Msg() or UserErrMsg() or Error()
	Params []interface{}
	Detail fmt.Stringer
}

func (GKEError) Error() string {
	panic("implement me")
}
