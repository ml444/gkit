package templates

var JsonImports = []string{
	"fmt",
	"encoding/json",
	"database/sql/driver",

	"google.golang.org/protobuf/proto",
	"google.golang.org/protobuf/encoding/protojson",
}

const JsonUtils = `
func jsonMarshal(x interface{}) ([]byte, error) {
	if m, ok := x.(proto.Message); ok {
		return protojson.Marshal(m)
	} 
	return json.Marshal(x)
}

func jsonUnmarshal(buf []byte, x interface{}) error {
	if m, ok := x.(proto.Message); ok {
		return protojson.Unmarshal(buf, m)
	} 
	return json.Unmarshal(buf, x)
}
`

const JsonTmpl = `
{{ if not .IsIgnore }}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return jsonUnmarshal(buf, x)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	b, err := jsonMarshal(&x)
	return string(b), err
}
{{ end }}
`

var SpecialJsonImports = []string{
	"fmt",
	"encoding/json",
	"database/sql/driver",
}

const SpecialJsonTmpl = `
type {{ .SerializerTypeName }} {{ .FieldType }}

{{ if not .IsIgnore -}}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return json.Unmarshal(buf, &x)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	b, err := json.Marshal(&x)
	return string(b), err
}
{{ end }}
`
