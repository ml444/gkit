package stream

const Name = "octet-stream"

func GetCoder() Coder {
	return Coder{}
}

type Coder struct{}

func (Coder) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (Coder) Unmarshal(data []byte, v interface{}) error {
	*(v.(*[]byte)) = data
	return nil
}

func (Coder) Name() string {
	return Name
}
