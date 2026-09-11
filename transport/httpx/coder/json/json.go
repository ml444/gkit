package json

import (
	"encoding/json"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the json Coder.
const Name = "json"

var (
	// MarshalOptions configures the legacy coder. Set it only during initialization,
	// before concurrent use. Prefer NewCoder for per-instance configuration.
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	// UnmarshalOptions configures the legacy coder. Set it only during initialization,
	// before concurrent use. Prefer NewCoder for per-instance configuration.
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func GetCoder() Coder {
	return Coder{}
}

// Coder is a Coder implementation with json.
type Coder struct{}

func (Coder) Marshal(v interface{}) ([]byte, error) {
	return marshal(v, MarshalOptions)
}

func marshal(v interface{}, opts protojson.MarshalOptions) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	case proto.Message:
		return opts.Marshal(m)
	default:
		return json.Marshal(m)
	}
}

func (Coder) Unmarshal(data []byte, v interface{}) error {
	return unmarshal(data, v, UnmarshalOptions)
}

func unmarshal(data []byte, v interface{}, opts protojson.UnmarshalOptions) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return json.Unmarshal(data, v)
	}
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return opts.Unmarshal(data, m)
	default:
		if rv.Kind() != reflect.Ptr {
			return json.Unmarshal(data, v)
		}
		for rv := rv; rv.Kind() == reflect.Ptr; {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
			return opts.Unmarshal(data, m)
		}
		return json.Unmarshal(data, m)
	}
}

func (Coder) Name() string {
	return Name
}
