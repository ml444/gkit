package funcs

import (
	"fmt"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
	"google.golang.org/protobuf/proto"
)

const (
	httpHeaderNamePattern  = `^:?[0-9a-zA-Z!#$%&'*+-.^_|~` + "`" + `]+$`
	httpHeaderValuePattern = `^[^\x00-\x08\x0A-\x1F\x7F]*$`
	httpHeaderLoosePattern = `^[^\x00\x0A\x0D]*$`
)

func headerStrict(rules *v.StringRules) bool {
	if rules == nil || rules.Strict == nil {
		return true
	}
	return rules.GetStrict()
}

func wellKnownRegexPattern(kr v.KnownRegex, strict bool) string {
	if !strict && (kr == v.KnownRegex_HTTP_HEADER_NAME || kr == v.KnownRegex_HTTP_HEADER_VALUE) {
		return httpHeaderLoosePattern
	}
	switch kr {
	case v.KnownRegex_HTTP_HEADER_NAME:
		return httpHeaderNamePattern
	case v.KnownRegex_HTTP_HEADER_VALUE:
		return httpHeaderValuePattern
	default:
		return ""
	}
}

func wellKnownRegexMessage(kr v.KnownRegex, strict bool) string {
	if !strict && (kr == v.KnownRegex_HTTP_HEADER_NAME || kr == v.KnownRegex_HTTP_HEADER_VALUE) {
		return "value contains invalid header characters"
	}
	switch kr {
	case v.KnownRegex_HTTP_HEADER_NAME:
		return "value must be a valid HTTP header name"
	case v.KnownRegex_HTTP_HEADER_VALUE:
		return "value must be a valid HTTP header value"
	default:
		return "value does not match the required well-known format"
	}
}

func stringRules(msg proto.Message) (*v.StringRules, bool) {
	r, ok := msg.(*v.StringRules)
	return r, ok && r != nil
}

// HasWellKnownRegex reports whether string rules use a known HTTP header regex.
func HasWellKnownRegex(msg proto.Message) bool {
	r, ok := stringRules(msg)
	if !ok {
		return false
	}
	return r.GetWellKnownRegex() != v.KnownRegex_UNKNOWN
}

// WellKnownRegexLit returns a quoted regex pattern for regexp.MustCompile in generated code.
func WellKnownRegexLit(msg proto.Message) string {
	r, ok := stringRules(msg)
	if !ok {
		return `""`
	}
	pat := wellKnownRegexPattern(r.GetWellKnownRegex(), headerStrict(r))
	return fmt.Sprintf("%q", pat)
}

// WellKnownRegexErrMsg returns the validation error message for a well-known regex rule.
func WellKnownRegexErrMsg(msg proto.Message) string {
	r, ok := stringRules(msg)
	if !ok {
		return "value does not match the required well-known format"
	}
	return wellKnownRegexMessage(r.GetWellKnownRegex(), headerStrict(r))
}
