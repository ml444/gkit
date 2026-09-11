package json

import "google.golang.org/protobuf/encoding/protojson"

// Options configures Protobuf JSON encoding. Ordinary Go values still use
// encoding/json, and custom JSON marshalers/unmarshalers retain precedence.
// Start with DefaultOptions to inherit the legacy defaults.
type Options struct {
	Marshal   protojson.MarshalOptions
	Unmarshal protojson.UnmarshalOptions
}

// DefaultOptions copies the legacy global options. Call during initialization;
// concurrent writes to the exported globals are not supported.
func DefaultOptions() Options {
	return Options{Marshal: MarshalOptions, Unmarshal: UnmarshalOptions}
}

// ConfiguredCoder is a JSON codec with immutable, per-instance options.
// Resolver objects referenced by those options are shared, not deep-copied,
// and must be safe for concurrent use.
type ConfiguredCoder struct {
	opts Options
}

// NewCoder copies opts without applying implicit defaults. Options{} therefore
// uses protojson's zero-value behavior; use DefaultOptions for legacy behavior.
func NewCoder(opts Options) *ConfiguredCoder {
	return &ConfiguredCoder{opts: opts}
}

func (c *ConfiguredCoder) Marshal(v interface{}) ([]byte, error) {
	return marshal(v, c.opts.Marshal)
}

func (c *ConfiguredCoder) Unmarshal(data []byte, v interface{}) error {
	return unmarshal(data, v, c.opts.Unmarshal)
}

func (*ConfiguredCoder) Name() string { return Name }
