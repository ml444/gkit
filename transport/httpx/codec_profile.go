package httpx

import (
	"github.com/ml444/gkit/transport/httpx/coder"
	jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
)

// codecProfile is built once for an explicitly configured router coder.
// A nil profile keeps the legacy dynamic registry lookup behavior.
type codecProfile struct {
	codecs map[string]coder.ICoder
}

func defaultCodecProfile() *codecProfile {
	codecs := coder.Snapshot()
	switch codecs[jsoncodec.Name].(type) {
	case jsoncodec.Coder, *jsoncodec.Coder:
		// Copying the registry alone would retain the legacy global JSON options.
		codecs[jsoncodec.Name] = jsoncodec.NewCoder(jsoncodec.DefaultOptions())
	}
	return &codecProfile{codecs: codecs}
}

func (p *codecProfile) get(name string) coder.ICoder {
	if p == nil {
		return coder.GetCoder(name)
	}
	if c, ok := p.codecs[name]; ok {
		return c
	}
	return p.codecs[jsoncodec.Name]
}
