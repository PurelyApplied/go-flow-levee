// Package proto fakes google3/net/proto2/go/proto package.
package proto

import (
	"fmt"
	"io"
	"reflect"
)

// Message fakes proto.Message.
type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

// Clone fakes proto.Clone.
func Clone(src Message) Message {
	in := reflect.ValueOf(src)
	if in.IsNil() {
		return src
	}
	out := reflect.New(in.Type().Elem())
	dst := out.Interface().(Message)
	// This is good enough for our purposes.
	return dst
}

// MarshalTextString fakes proto.MarshalTestString.
func MarshalTextString(pb Message) string {
	return "Some text"
}

// MarshalText fakes proto.MarshalText
func MarshalText(w io.Writer, pb Message) error {
	fmt.Fprint(w, "Some text")
	return nil
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	return &v
}
