package encoding

import (
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

var registeredCodecs = make(map[string]Coder)

// RegisterCodec registers the provided Coder for use with all Transport clients and
// servers.
func RegisterCodec(codec Coder) {
	if codec == nil {
		panic("cannot register a nil Coder")
	}
	if codec.Name() == "" {
		panic("cannot register Coder with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(codec.Name())
	registeredCodecs[contentSubtype] = codec
}

// GetCoder gets a registered Coder by content-subtype, or nil if no Coder is
// registered for the content-subtype.
//
// The content-subtype is expected to be lowercase.
func GetCoder(contentSubtype string) Coder {
	return registeredCodecs[contentSubtype]
}
