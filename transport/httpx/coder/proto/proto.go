// Package proto defines the protobuf Coder. Importing this package will
// register the Coder.
package proto

import (
	"errors"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the proto compressor.
const Name = "proto"

func GetCoder() Coder {
	return Coder{}
}

// Coder is a Coder implementation with protobuf. It is the default Coder for Transport.
type Coder struct{}

func (Coder) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (Coder) Unmarshal(data []byte, v interface{}) error {
	pm, err := getProtoMessage(v)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, pm)
}

func (Coder) Name() string {
	return Name
}

func getProtoMessage(v interface{}) (proto.Message, error) {
	if msg, ok := v.(proto.Message); ok {
		return msg, nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("not proto message")
	}

	val = val.Elem()
	return getProtoMessage(val.Interface())
}
