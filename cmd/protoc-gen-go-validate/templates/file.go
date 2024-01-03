package templates

const FileTmpl = `import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	//"google.golang.org/protobuf/types/known/anypb"

	{{ range $pkg, $pkgPath := GetImports . }}
		{{ $pkg }} "{{ $pkgPath }}"
	{{ end }}
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	//_ = anypb.Any{}
	_ = sort.Sort

	{{ range $pkg, $enum := enumPackages (externalEnums .Desc) }}
	_ = {{ $pkg }}.{{ $enum.Name }}(0)
	{{ end }}
)

{{- if fileneeds . "uuid" }}
// define the regex for a UUID once up-front
var _uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

{{ template "uuid" . }}
{{ end }}

{{- $notExistValidateHost := true }}
{{- if fileneeds . "hostname" }}
{{- $notExistValidateHost = false }}
{{ template "hostname" . }}
{{ end }}


{{- if fileneeds . "email" }}
{{ template "email" . }}
{{- if $notExistValidateHost }}
{{ template "hostname" . }}
{{ end }}
{{ end }}
`

const CommonDefTmpl = `
// ValidationError is the validation error 
// returned by .Validate if the designated constraints aren't met.
type ValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e  ValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e  ValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e  ValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e  ValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e  ValidationError) ErrorName() string { return "ValidationError" }

// Error satisfies the builtin error interface
func (e  ValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %s .%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

// ValidationError is an error wrapping multiple validation errors 
// returned by .ValidateAll() if the designated constraints aren't met.
type  MultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MultiError) AllErrors() []error { return m }
`
