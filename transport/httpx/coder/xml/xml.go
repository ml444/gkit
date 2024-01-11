package xml

import (
	"encoding/xml"
)

// Name is the name registered for the xml Coder.
const Name = "xml"

func GetCoder() Coder {
	return Coder{}
}

// Coder is a Coder implementation with xml.
type Coder struct{}

func (Coder) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (Coder) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (Coder) Name() string {
	return Name
}
