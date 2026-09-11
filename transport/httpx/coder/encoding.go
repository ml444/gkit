package coder

import (
	"errors"
	"maps"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

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

// Published snapshots are immutable. Readers load one version without taking
// a mutex; writers serialize the entire copy-and-publish operation.
type registrySnapshot struct {
	codecs map[string]ICoder
}

var registryWriteMu sync.Mutex
var registeredCoders atomic.Pointer[registrySnapshot]

func init() {
	registeredCoders.Store(&registrySnapshot{codecs: map[string]ICoder{
		xml.Name:    xml.GetCoder(),
		form.Name:   form.GetCoder(),
		json.Name:   json.GetCoder(),
		eproto.Name: eproto.GetCoder(),
		stream.Name: stream.GetCoder(),
	}})
}

// RegisterCoder registers the provided ICoder for use with all Transport clients and servers.
func RegisterCoder(codec ICoder) error {
	if codec == nil || isNilCoder(codec) {
		return errors.New("cannot register a nil ICoder")
	}
	name := codec.Name()
	if name == "" {
		return errors.New("cannot register ICoder with empty string result for Name()")
	}
	contentSubtype := strings.ToLower(name)
	registryWriteMu.Lock()
	defer registryWriteMu.Unlock()
	next := maps.Clone(registeredCoders.Load().codecs)
	next[contentSubtype] = codec
	registeredCoders.Store(&registrySnapshot{codecs: next})
	return nil
}

// GetCoder gets a registered ICoder by content-subtype
// The content-subtype is expected to be lowercase.
func GetCoder(contentSubtype string) ICoder {
	codecs := registeredCoders.Load().codecs
	c, ok := codecs[contentSubtype]
	if !ok {
		return codecs[json.Name]
	}
	return c
}

// LookupCoder looks up a codec name case-insensitively without falling back to JSON.
func LookupCoder(name string) (ICoder, bool) {
	c, ok := registeredCoders.Load().codecs[strings.ToLower(name)]
	return c, ok
}

// Snapshot returns an independent copy of one published registry version. Codec objects
// are shared and must remain safe for concurrent use; only the map is copied.
func Snapshot() map[string]ICoder {
	return maps.Clone(registeredCoders.Load().codecs)
}

func isNilCoder(codec ICoder) bool {
	v := reflect.ValueOf(codec)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
