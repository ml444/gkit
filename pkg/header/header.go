package header

import "strings"

type IHeader interface {
	Get(key string) string
	Set(key string, value string)
	Append(key string, value string)
	Keys() []string
	Values(key string) []string
}
type Header map[string][]string

// New creates an MD from a given key-values map.
func New(mds ...map[string][]string) Header {
	md := Header{}
	for _, m := range mds {
		for k, vList := range m {
			for _, v := range vList {
				md.Append(k, v)
			}
		}
	}
	return md
}

// Append adds the key, value pair to the header.
func (m Header) Append(key, value string) {
	if len(key) == 0 {
		return
	}

	m[strings.ToLower(key)] = append(m[strings.ToLower(key)], value)
}

// Get returns the value associated with the passed key.
func (m Header) Get(key string) string {
	v := m[strings.ToLower(key)]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// Set stores the key-value pair.
func (m Header) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}
	m[strings.ToLower(key)] = []string{value}
}

// Range iterate over element in metadata.
func (m Header) Range(f func(k string, v []string) bool) {
	for k, v := range m {
		if !f(k, v) {
			break
		}
	}
}

func (m Header) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of values associated with the passed key.
func (m Header) Values(key string) []string {
	return m[strings.ToLower(key)]
}

// Clone returns a deep copy of Header
func (m Header) Clone() Header {
	md := make(Header, len(m))
	for k, v := range m {
		md[k] = v
	}
	return md
}
