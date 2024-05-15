// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v4.25.0--rc2
// source: cases/strings.proto

package cases

import (
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
	// "google.golang.org/protobuf/types/known/anypb"
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
)

// define the regex for a UUID once up-front
var _uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

func _validateUuid(uuid string) error {
	if matched := _uuidPattern.MatchString(uuid); !matched {
		return errors.New("invalid uuid format")
	}

	return nil
}

func _validateHostname(host string) error {
	s := strings.ToLower(strings.TrimSuffix(host, "."))

	if len(host) > 253 {
		return errors.New("hostname cannot exceed 253 characters")
	}

	for _, part := range strings.Split(s, ".") {
		if l := len(part); l == 0 || l > 63 {
			return errors.New("hostname part must be non-empty and cannot exceed 63 characters")
		}

		if part[0] == '-' {
			return errors.New("hostname parts cannot begin with hyphens")
		}

		if part[len(part)-1] == '-' {
			return errors.New("hostname parts cannot end with hyphens")
		}

		for _, r := range part {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
				return fmt.Errorf("hostname parts can only contain alphanumeric characters or hyphens, got %q", string(r))
			}
		}
	}

	return nil
}

func _validateEmail(addr string) error {
	a, err := mail.ParseAddress(addr)
	if err != nil {
		return err
	}
	addr = a.Address

	if len(addr) > 254 {
		return errors.New("email addresses cannot exceed 254 characters")
	}

	parts := strings.SplitN(addr, "@", 2)

	if len(parts[0]) > 64 {
		return errors.New("email address local phrase cannot exceed 64 characters")
	}

	return _validateHostname(parts[1])
}

// Validate checks the field values on StringConst with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringConst) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringConst with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringConstMultiError, or
// nil if none found.
func (m *StringConst) ValidateAll() error {
	return m.validate(true)
}

func (m *StringConst) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetVal() != "foo" {
		err := ValidationError{
			field:   "Val",
			reason:  "value must equal foo",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringIn with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringIn with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringInMultiError, or nil
// if none found.
func (m *StringIn) ValidateAll() error {
	return m.validate(true)
}

func (m *StringIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if _, ok := _StringIn_Val_InLookup[m.GetVal()]; !ok {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be in list [bar baz]",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

var _StringIn_Val_InLookup = map[string]struct{}{
	"bar": {},
	"baz": {},
}

// Validate checks the field values on StringNotIn with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringNotIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringNotIn with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringNotInMultiError, or
// nil if none found.
func (m *StringNotIn) ValidateAll() error {
	return m.validate(true)
}

func (m *StringNotIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if _, ok := _StringNotIn_Val_NotInLookup[m.GetVal()]; ok {
		err := ValidationError{
			field:   "Val",
			reason:  "value must not be in list [fizz buzz]",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

var _StringNotIn_Val_NotInLookup = map[string]struct{}{
	"fizz": {},
	"buzz": {},
}

// Validate checks the field values on StringLen with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringLen with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringLenMultiError, or nil
// if none found.
func (m *StringLen) ValidateAll() error {
	return m.validate(true)
}

func (m *StringLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetVal()) != 3 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be 3 runes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMinLen with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringMinLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMinLen with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringMinLenMultiError, or
// nil if none found.
func (m *StringMinLen) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMinLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetVal()) < 3 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at least 3 runes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMaxLen with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMaxLen with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringMaxLenMultiError, or
// nil if none found.
func (m *StringMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetVal()) > 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at most 5 runes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMinMaxLen with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *StringMinMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMinMaxLen with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringMinMaxLenMultiError, or nil if none found.
func (m *StringMinMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMinMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetVal()); l < 3 || l > 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be between 3 and 5 runes, inclusive",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringEqualMinMaxLen with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *StringEqualMinMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringEqualMinMaxLen with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringEqualMinMaxLenMultiError, or nil if none found.
func (m *StringEqualMinMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *StringEqualMinMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetVal()) != 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be 5 runes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)

	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringLenBytes with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringLenBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringLenBytes with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringLenBytesMultiError,
// or nil if none found.
func (m *StringLenBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringLenBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) != 4 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be 4 bytes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMinBytes with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringMinBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMinBytes with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringMinBytesMultiError,
// or nil if none found.
func (m *StringMinBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMinBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) < 4 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at least 4 bytes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMaxBytes with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringMaxBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMaxBytes with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringMaxBytesMultiError,
// or nil if none found.
func (m *StringMaxBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMaxBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) > 8 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at most 8 bytes",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringMinMaxBytes with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *StringMinMaxBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringMinMaxBytes with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringMinMaxBytesMultiError, or nil if none found.
func (m *StringMinMaxBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringMinMaxBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetVal()); l < 4 || l > 8 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be between 4 and 8 bytes, inclusive",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringEqualMinMaxBytes with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *StringEqualMinMaxBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringEqualMinMaxBytes with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringEqualMinMaxBytesMultiError, or nil if none found.
func (m *StringEqualMinMaxBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringEqualMinMaxBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetVal()); l < 4 || l > 8 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be between 4 and 8 bytes, inclusive",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringPattern with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringPattern) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringPattern with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringPatternMultiError, or
// nil if none found.
func (m *StringPattern) ValidateAll() error {
	return m.validate(true)
}

func (m *StringPattern) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !_StringPattern_Val_Pattern.MatchString(m.GetVal()) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not match regex pattern \"(?i)^[a-z0-9]+$\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

var _StringPattern_Val_Pattern = regexp.MustCompile("(?i)^[a-z0-9]+$")

// Validate checks the field values on StringPatternEscapes with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *StringPatternEscapes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringPatternEscapes with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringPatternEscapesMultiError, or nil if none found.
func (m *StringPatternEscapes) ValidateAll() error {
	return m.validate(true)
}

func (m *StringPatternEscapes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !_StringPatternEscapes_Val_Pattern.MatchString(m.GetVal()) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not match regex pattern \"\\\\* \\\\\\\\ \\\\w\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

var _StringPatternEscapes_Val_Pattern = regexp.MustCompile("\\* \\\\ \\w")

// Validate checks the field values on StringPrefix with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringPrefix) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringPrefix with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringPrefixMultiError, or
// nil if none found.
func (m *StringPrefix) ValidateAll() error {
	return m.validate(true)
}

func (m *StringPrefix) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !strings.HasPrefix(m.GetVal(), "foo") {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not have prefix \"foo\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringContains with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringContains) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringContains with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringContainsMultiError,
// or nil if none found.
func (m *StringContains) ValidateAll() error {
	return m.validate(true)
}

func (m *StringContains) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !strings.Contains(m.GetVal(), "bar") {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not contain substring \"bar\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringNotContains with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *StringNotContains) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringNotContains with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringNotContainsMultiError, or nil if none found.
func (m *StringNotContains) ValidateAll() error {
	return m.validate(true)
}

func (m *StringNotContains) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if strings.Contains(m.GetVal(), "bar") {
		err := ValidationError{
			field:   "Val",
			reason:  "value contains substring \"bar\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringSuffix with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringSuffix) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringSuffix with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringSuffixMultiError, or
// nil if none found.
func (m *StringSuffix) ValidateAll() error {
	return m.validate(true)
}

func (m *StringSuffix) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !strings.HasSuffix(m.GetVal(), "baz") {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not have suffix \"baz\"",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringEmail with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringEmail) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringEmail with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringEmailMultiError, or
// nil if none found.
func (m *StringEmail) ValidateAll() error {
	return m.validate(true)
}

func (m *StringEmail) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if err := _validateEmail(m.GetVal()); err != nil {
		err = ValidationError{
			field:   "Val",
			reason:  "value must be a valid email address",
			errCode: 0,
			cause:   err,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringAddress with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringAddress) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringAddress with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringAddressMultiError, or
// nil if none found.
func (m *StringAddress) ValidateAll() error {
	return m.validate(true)
}

func (m *StringAddress) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if err := _validateHostname(m.GetVal()); err != nil {
		if ip := net.ParseIP(m.GetVal()); ip == nil {
			err := ValidationError{
				field:   "Val",
				reason:  "value must be a valid hostname, or ip address",
				errCode: 0,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringHostname with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringHostname) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringHostname with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringHostnameMultiError,
// or nil if none found.
func (m *StringHostname) ValidateAll() error {
	return m.validate(true)
}

func (m *StringHostname) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if err := _validateHostname(m.GetVal()); err != nil {
		err = ValidationError{
			field:   "Val",
			reason:  "value must be a valid hostname",
			errCode: 0,
			cause:   err,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringIP with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringIP) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringIP with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringIPMultiError, or nil
// if none found.
func (m *StringIP) ValidateAll() error {
	return m.validate(true)
}

func (m *StringIP) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.ParseIP(m.GetVal()); ip == nil {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be a valid IP address",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringIPv4 with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringIPv4) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringIPv4 with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringIPv4MultiError, or
// nil if none found.
func (m *StringIPv4) ValidateAll() error {
	return m.validate(true)
}

func (m *StringIPv4) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.ParseIP(m.GetVal()); ip == nil || ip.To4() == nil {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be a valid IPv4 address",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringIPv6 with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringIPv6) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringIPv6 with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringIPv6MultiError, or
// nil if none found.
func (m *StringIPv6) ValidateAll() error {
	return m.validate(true)
}

func (m *StringIPv6) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.ParseIP(m.GetVal()); ip == nil || ip.To4() != nil {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be a valid IPv6 address",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringURI with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringURI) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringURI with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringURIMultiError, or nil
// if none found.
func (m *StringURI) ValidateAll() error {
	return m.validate(true)
}

func (m *StringURI) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if uri, err := url.Parse(m.GetVal()); err != nil {
		err = ValidationError{
			field:   "Val",
			reason:  "value must be a valid URI",
			errCode: 0,
			cause:   err,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	} else if !uri.IsAbs() {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be absolute",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringURIRef with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringURIRef) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringURIRef with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringURIRefMultiError, or
// nil if none found.
func (m *StringURIRef) ValidateAll() error {
	return m.validate(true)
}

func (m *StringURIRef) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if _, err := url.Parse(m.GetVal()); err != nil {
		err = ValidationError{
			field:   "Val",
			reason:  "value must be a valid URI",
			errCode: 0,
			cause:   err,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringUUID with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringUUID) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringUUID with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringUUIDMultiError, or
// nil if none found.
func (m *StringUUID) ValidateAll() error {
	return m.validate(true)
}

func (m *StringUUID) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if err := _validateUuid(m.GetVal()); err != nil {
		err = ValidationError{
			field:   "Val",
			reason:  "value must be a valid UUID",
			errCode: 0,
			cause:   err,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringHttpHeaderName with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *StringHttpHeaderName) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringHttpHeaderName with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringHttpHeaderNameMultiError, or nil if none found.
func (m *StringHttpHeaderName) ValidateAll() error {
	return m.validate(true)
}

func (m *StringHttpHeaderName) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringHttpHeaderValue with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *StringHttpHeaderValue) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringHttpHeaderValue with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringHttpHeaderValueMultiError, or nil if none found.
func (m *StringHttpHeaderValue) ValidateAll() error {
	return m.validate(true)
}

func (m *StringHttpHeaderValue) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringValidHeader with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *StringValidHeader) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringValidHeader with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringValidHeaderMultiError, or nil if none found.
func (m *StringValidHeader) ValidateAll() error {
	return m.validate(true)
}

func (m *StringValidHeader) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringUUIDIgnore with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *StringUUIDIgnore) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringUUIDIgnore with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// StringUUIDIgnoreMultiError, or nil if none found.
func (m *StringUUIDIgnore) ValidateAll() error {
	return m.validate(true)
}

func (m *StringUUIDIgnore) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetVal() != "" {

		if err := _validateUuid(m.GetVal()); err != nil {
			err = ValidationError{
				field:   "Val",
				reason:  "value must be a valid UUID",
				errCode: 0,
				cause:   err,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on StringInOneOf with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *StringInOneOf) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on StringInOneOf with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in StringInOneOfMultiError, or
// nil if none found.
func (m *StringInOneOf) ValidateAll() error {
	return m.validate(true)
}

func (m *StringInOneOf) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	switch v := m.Foo.(type) {
	case *StringInOneOf_Bar:
		if v == nil {
			err := ValidationError{
				field:  "Foo",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if _, ok := _StringInOneOf_Bar_InLookup[m.GetBar()]; !ok {
			err := ValidationError{
				field:   "Bar",
				reason:  "value must be in list [a b]",
				errCode: 0,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

var _StringInOneOf_Bar_InLookup = map[string]struct{}{
	"a": {},
	"b": {},
}
