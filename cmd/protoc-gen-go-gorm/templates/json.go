package templates

var JsonImports = []string{
	"fmt",
	"encoding/json",
	"database/sql/driver",

	"google.golang.org/protobuf/proto",
	"google.golang.org/protobuf/encoding/protojson",
}

const JsonTmpl = `
{{ if not .IsIgnore }}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	var data []byte
	switch buf := src.(type) {
	case []byte:
		data = buf
	case string:
		data = []byte(buf)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
	if message, ok := interface{}(x).(proto.Message); ok {
		return protojson.Unmarshal(data, message)
	}
	return json.Unmarshal(data, x)
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	if message, ok := interface{}(&x).(proto.Message); ok {
		b, err := protojson.Marshal(message)
		return string(b), err
	}
	b, err := json.Marshal(&x)
	return string(b), err
}
{{ end }}
`

var ExternalJsonImports = []string{
	"fmt",
	"database/sql/driver",

	"google.golang.org/protobuf/encoding/protojson",
}

const ExternalJsonTmpl = `
type {{ .SerializerTypeName }} {{ .FieldType }}

{{ if not .IsIgnore -}}
func (x *{{ .SerializerTypeName }}) Scan(src interface{}) error {
	var data []byte
	switch buf := src.(type) {
	case []byte:
		data = buf
	case string:
		data = []byte(buf)
	default:
		return fmt.Errorf("{{ .SerializerTypeName }} unsupported type [%s] to scan", buf)
	}
	return protojson.Unmarshal(data, (*{{ .FieldType }})(x))
}

func (x {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	b, err := protojson.Marshal((*{{ .FieldType }})(&x))
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
	case string:
		return json.Unmarshal([]byte(buf), &x)
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
