package transport

import (
	"fmt"
	"strings"
)

type MD map[string][]string

// New creates an MD from a given key-values map.
func New(mds ...map[string][]string) MD {
	md := MD{}
	for _, m := range mds {
		for k, vList := range m {
			md.Append(k, vList...)
		}
	}
	return md
}

// Pairs returns an MD formed by the mapping of key, value ...
// Pairs panics if len(kv) is odd.
func Pairs(kv ...string) MD {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kv)))
	}
	md := make(MD, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		md.Append(kv[i], kv[i+1])
	}
	return md
}

// Append adds the values to the key.
func (m MD) Append(key string, values ...string) {
	if len(values) == 0 {
		return
	}
	key = m.normalizeKey(key)
	m[key] = append(m[key], values...)
}

// GetFirst obtains the first value for a given key.
func (m MD) GetFirst(key string) string {
	v := m.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// Get obtains the values for a given key.
func (m MD) Get(key string) []string {
	key = m.normalizeKey(key)
	return m[key]
}

// Set stores the key-value pair.
func (m MD) Set(key string, values ...string) {
	if len(values) == 0 {
		return
	}
	key = m.normalizeKey(key)
	m[key] = values
}

// Delete removes the values for a given key.
func (m MD) Delete(k string) {
	k = m.normalizeKey(k)
	delete(m, k)
	for key := range m {
		if strings.EqualFold(key, k) {
			delete(m, key)
		}
	}
}

// Range iterate over element in metadata.
func (m MD) Range(f func(k string, v []string) bool) {
	for k, v := range m.Copy() {
		if !f(k, v) {
			break
		}
	}
}

func (m MD) Keys() []string {
	var keys []string
	for k := range m.Copy() {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of items in md.
func (m MD) Len() int {
	return len(m.Copy())
}

// Copy returns a deep copy of MD
func (m MD) Copy() MD {
	md := make(MD, len(m))
	for k, v := range m {
		md.Append(k, v...)
	}
	return md
}

// Merge joins any number of mds into a single MD.
func Merge(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out.Append(k, v...)
		}
	}
	return out
}

func (m MD) normalizeKey(key string) string {
	return strings.ToLower(key)
}


