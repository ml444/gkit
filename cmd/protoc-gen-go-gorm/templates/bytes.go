package templates

var BytesImports = []string{
	"fmt",
	"database/sql/driver",

	"google.golang.org/protobuf/proto",
}

const BytesUtils = `
func bytesMarshal(x interface{}) ([]byte, error) {
	if m, ok := x.(proto.Message); ok {
		return proto.Marshal(m)
	} 
	return json.Marshal(x)
}

func bytesUnmarshal(buf []byte, x interface{}) error {
	if m, ok := x.(proto.Message); ok {
		return proto.Unmarshal(buf, m)
	} 
	return json.Unmarshal(buf, x)
}
`
const BytesTmpl = `
{{ if not .IsIgnore -}}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return bytesUnmarshal(buf, x)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	b, err := bytesMarshal(&x)
	return b, err
}
{{ end }}
`

var SpecialBytesImports = []string{
	"fmt",
	"encoding/json",
	"database/sql/driver",
}

const SpecialBytesTmpl = `
type {{ .SerializerTypeName }} {{ .FieldType }}
{{ if not .IsIgnore }}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return bytesUnmarshal(buf, &x)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	b, err := bytesMarshal(&x)
	return b, err
}
{{ end }}
`
