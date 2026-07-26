package stream

import "fmt"

const Name = "octet-stream"

func GetCoder() Coder {
	return Coder{}
}

type Coder struct{}

func (Coder) Marshal(v interface{}) ([]byte, error) {
	if vv, ok := v.([]byte); ok {
		return vv, nil
	}
	return nil, fmt.Errorf("not stream type: %T", v)
}

func (Coder) Unmarshal(data []byte, v interface{}) error {
	*(v.(*[]byte)) = data
	return nil
}

func (Coder) Name() string {
	return Name
}
