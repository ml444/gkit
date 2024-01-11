package form

import (
	"net/url"
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/go-playground/form"
)

const (
	// Name is form Coder name
	Name = "x-www-form-urlencoded"
	// Null value string
	nullStr = "null"
)

var (
	encoder = form.NewEncoder()
	decoder = form.NewDecoder()
)

func GetCoder() Coder {
	decoder.SetTagName("json")
	encoder.SetTagName("json")
	return Coder{encoder: encoder, decoder: decoder}
}

type Coder struct {
	encoder *form.Encoder
	decoder *form.Decoder
}

func (c Coder) Marshal(v interface{}) ([]byte, error) {
	var vs url.Values
	var err error
	if m, ok := v.(proto.Message); ok {
		vs, err = EncodeValues(m)
		if err != nil {
			return nil, err
		}
	} else {
		vs, err = c.encoder.Encode(v)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range vs {
		if len(v) == 0 {
			delete(vs, k)
		}
	}
	return []byte(vs.Encode()), nil
}

func (c Coder) Unmarshal(data []byte, v interface{}) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, vs)
	}
	if m, ok := rv.Interface().(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	return c.decoder.Decode(v, vs)
}

func (Coder) Name() string {
	return Name
}
