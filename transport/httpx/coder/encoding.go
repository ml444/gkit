package coder

import (
	"errors"
	"github.com/ml444/gkit/transport/httpx/coder/form"
	"github.com/ml444/gkit/transport/httpx/coder/json"
	eproto "github.com/ml444/gkit/transport/httpx/coder/proto"
	"github.com/ml444/gkit/transport/httpx/coder/xml"
	"strings"
)

// Coder defines the interface Transport uses to encode and decode messages.  Note
// that implementations of this interface must be thread safe; a Coder's
// methods can be called from concurrent goroutines.
type Coder interface {
	// Marshal returns the wire format of v.
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal parses the wire format into v.
	Unmarshal(data []byte, v interface{}) error
	// Name returns the name of the Coder implementation. The returned string
	// will be used as part of content type in transmission.  The result must be
	// static; the result cannot change between calls.
	Name() string
}

var registeredCoders = map[string]Coder{
	form.Name:   form.GetCoder(),
	json.Name:   json.GetCoder(),
	eproto.Name: eproto.GetCoder(),
	xml.Name:    xml.GetCoder(),
}

// RegisterCoder registers the provided Coder for use with all Transport clients and servers.
func RegisterCoder(codec Coder) error {
	if codec == nil {
		return errors.New("cannot register a nil Coder")
	}
	if codec.Name() == "" {
		return errors.New("cannot register Coder with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCoders[contentSubtype] = codec
	return nil
}

// GetCoder gets a registered Coder by content-subtype
// The content-subtype is expected to be lowercase.
func GetCoder(contentSubtype string) Coder {
	return registeredCoders[contentSubtype]
}
