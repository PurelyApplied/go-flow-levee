// Package status fakes cloud/util/task/go/status
package status

import (
	"errors"
	"fmt"
)

// Code fakes status.Code
type Code int

// FakeCode fakes an error code
const FakeCode = Code(1234)

// Errorf fakes status.Errorf
func Errorf(c Code, format string, a ...interface{}) error {
	return errors.New("fake error")
}

// Error fakes status.Error
func Error(code Code, msg string) error {
	return fmt.Errorf("faked error, code %d, msg %q", code, msg)
}
