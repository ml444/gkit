package coder

import (
	"errors"
	"strings"

	"github.com/ml444/gkit/transport/httpx/coder/form"
	"github.com/ml444/gkit/transport/httpx/coder/json"
	eproto "github.com/ml444/gkit/transport/httpx/coder/proto"
	"github.com/ml444/gkit/transport/httpx/coder/stream"
	"github.com/ml444/gkit/transport/httpx/coder/xml"
)

// ICoder defines the interface Transport uses to encode and decode messages.  Note
// that implementations of this interface must be thread safe; a ICoder's
// methods can be called from concurrent goroutines.
type ICoder interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

var registeredCoders = map[string]ICoder{
	xml.Name:    xml.GetCoder(),
	form.Name:   form.GetCoder(),
	json.Name:   json.GetCoder(),
	eproto.Name: eproto.GetCoder(),
	stream.Name: stream.GetCoder(),
}

// RegisterCoder registers the provided ICoder for use with all Transport clients and servers.
func RegisterCoder(codec ICoder) error {
	if codec == nil {
		return errors.New("cannot register a nil ICoder")
	}
	if codec.Name() == "" {
		return errors.New("cannot register ICoder with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCoders[contentSubtype] = codec
	return nil
}

// GetCoder gets a registered ICoder by content-subtype
// The content-subtype is expected to be lowercase.
func GetCoder(contentSubtype string) ICoder {
	c, ok := registeredCoders[contentSubtype]
	if !ok {
		return registeredCoders[json.Name]
	}
	return c
}
