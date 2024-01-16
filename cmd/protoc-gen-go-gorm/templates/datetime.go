package templates

var DateImports = []string{
	"time",
	"fmt",
	"database/sql",
	"database/sql/driver",

	"google.golang.org/protobuf/types/known/durationpb",
	"google.golang.org/protobuf/types/known/timestamppb",
}

const DateUtils = `
const (
	DateTime   = "2006-01-02 15:04:05"
	DateOnly   = "2006-01-02"
	TimeOnly   = "15:04:05"
)

func scanDatetime(dt interface{}, t time.Time, layout string) (d interface{}, err error) {
	switch dt.(type) {
	case string:
		d = t.Format(layout)
	case int32:
		d = int32(t.Unix())
	case int64:
		d = t.Unix()
	case uint32:
		d = uint32(t.Unix())
	case uint64:
		d = uint64(t.Unix())
	case time.Time:
		d = t
	case *timestamppb.Timestamp:
		d = &timestamppb.Timestamp{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		}
	case *durationpb.Duration:
		d = &durationpb.Duration{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		}
	default:
		err = fmt.Errorf("conversion of [%T] type is not supported", dt)
	}
	return
}

func valueToTime(dt interface{}, layout string) (t time.Time, err error) {
	switch d := dt.(type) {
	case string:
		return time.Parse(layout, d)
	case int32:
		t = time.Unix(int64(d), 0)
	case int64:
		t = time.Unix(d, 0)
	case uint32:
		t = time.Unix(int64(d), 0)
	case uint64:
		t = time.Unix(int64(d), 0)
	case time.Time:
		t = d
	case *timestamppb.Timestamp:
		t = time.Unix(d.Seconds, int64(d.Nanos))
	case *durationpb.Duration:
		t = time.Unix(d.Seconds, int64(d.Nanos))
	default:
		err = fmt.Errorf("conversion of [%T] type is not supported", dt)
	}
	return
}
`

const DateTmpl = `
type {{ .SerializerTypeName }} {{ .FieldType }}

func (dt *{{ .SerializerTypeName }}) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	if err != nil {
		return err
	}
	var realTyp {{ .FieldType }}
	{{ if eq .SerializerName "date" }}
	val, err := scanDatetime(realTyp, nullTime.Time, DateOnly)
	{{ else if eq .SerializerName "time" -}}
	val, err := scanDatetime(realTyp, nullTime.Time, TimeOnly)
	{{ else -}}
	val, err := scanDatetime(realTyp, nullTime.Time, DateTime)
	{{ end -}}
	if err != nil {
		return fmt.Errorf("scan {{ .SerializerTypeName }} error: %w", err)
	}
	*dt = {{ .SerializerTypeName }}(val.({{ .FieldType }}))
	return
}

func (dt {{ .SerializerTypeName }}) Value() (driver.Value, error) {
	{{- if eq .SerializerName "date" -}}
	t, err := valueToTime({{ .FieldType }}(dt), DateOnly)
	{{ else if eq .SerializerName "time" -}}
	t, err := valueToTime({{ .FieldType }}(dt), TimeOnly)
	{{ else -}}
	t, err := valueToTime({{ .FieldType }}(dt), DateTime)
	{{ end -}}
	if err != nil {
		return nil, fmt.Errorf("value {{ .SerializerTypeName }} error: %w", err)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()), nil
}
`
