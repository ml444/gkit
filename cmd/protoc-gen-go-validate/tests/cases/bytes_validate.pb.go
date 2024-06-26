// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v4.25.0--rc2
// source: cases/bytes.proto

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

	"github.com/ml444/gkit/errorx"
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

// ValidationError is the validation error
// returned by .Validate if the designated constraints aren't met.
type ValidationError struct {
	field   string
	reason  string
	errCode int32
	cause   error
	key     bool
}

// Field function returns field value.
func (e ValidationError) Field() string { return e.field }

func (e ValidationError) Reason() string { return e.reason }

func (e ValidationError) Cause() error { return e.cause }

func (e ValidationError) Key() bool { return e.key }

func (e ValidationError) ErrorName() string { return "ValidationError" }

// Error satisfies the builtin error interface
func (e ValidationError) Error() string {
	if e.errCode != 0 {
		return errorx.New(e.errCode).Error()
	} else {
		cause := ""
		if e.cause != nil {
			cause = fmt.Sprintf(" | caused by: %v", e.cause)
		}

		key := ""
		if e.key {
			key = "key for "
		}

		msg := fmt.Sprintf("invalid %s .%s: %s%s", key, e.field, e.reason, cause)
		return errorx.CreateError(400, 0, msg).Error()
	}
}

// ValidationError is an error wrapping multiple validation errors
// returned by .ValidateAll() if the designated constraints aren't met.
type MultiError []error

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

// Validate checks the field values on BytesConst with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesConst) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesConst with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesConstMultiError, or
// nil if none found.
func (m *BytesConst) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesConst) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !bytes.Equal(m.GetVal(), []uint8{0x66, 0x6F, 0x6F}) {
		err := ValidationError{
			field:   "Val",
			reason:  "value must equal [102 111 111]",
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

// Validate checks the field values on BytesIn with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesIn with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in BytesInMultiError, or nil if none found.
func (m *BytesIn) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if _, ok := _BytesIn_Val_InLookup[string(m.GetVal())]; !ok {
		err := ValidationError{
			field:   "Val",
			reason:  "value must be in list [[98 97 114] [98 97 122]]",
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

var _BytesIn_Val_InLookup = map[string]struct{}{
	"\x62\x61\x72": {},
	"\x62\x61\x7A": {},
}

// Validate checks the field values on BytesNotIn with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesNotIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesNotIn with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesNotInMultiError, or
// nil if none found.
func (m *BytesNotIn) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesNotIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if _, ok := _BytesNotIn_Val_NotInLookup[string(m.GetVal())]; ok {
		err := ValidationError{
			field:   "Val",
			reason:  "value must not be in list [[102 105 122 122] [98 117 122 122]]",
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

var _BytesNotIn_Val_NotInLookup = map[string]struct{}{
	"\x66\x69\x7A\x7A": {},
	"\x62\x75\x7A\x7A": {},
}

// Validate checks the field values on BytesLen with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesLen with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesLenMultiError, or nil
// if none found.
func (m *BytesLen) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) != 3 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be 3 bytes",
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

// Validate checks the field values on BytesMinLen with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesMinLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesMinLen with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesMinLenMultiError, or
// nil if none found.
func (m *BytesMinLen) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesMinLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) < 3 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at least 3 bytes",
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

// Validate checks the field values on BytesMaxLen with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesMaxLen with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesMaxLenMultiError, or
// nil if none found.
func (m *BytesMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) > 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be at most 5 bytes",
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

// Validate checks the field values on BytesMinMaxLen with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesMinMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesMinMaxLen with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesMinMaxLenMultiError,
// or nil if none found.
func (m *BytesMinMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesMinMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := len(m.GetVal()); l < 3 || l > 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be between 3 and 5 bytes, inclusive",
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

// Validate checks the field values on BytesEqualMinMaxLen with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *BytesEqualMinMaxLen) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesEqualMinMaxLen with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// BytesEqualMinMaxLenMultiError, or nil if none found.
func (m *BytesEqualMinMaxLen) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesEqualMinMaxLen) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) != 5 {
		err := ValidationError{
			field:   "Val",
			reason:  "value length must be 5 bytes",
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

// Validate checks the field values on BytesPattern with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesPattern) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesPattern with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesPatternMultiError, or
// nil if none found.
func (m *BytesPattern) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesPattern) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !_BytesPattern_Val_Pattern.Match(m.GetVal()) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not match regex pattern \"^[\\x00-\\x7f]+$\"",
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

var _BytesPattern_Val_Pattern = regexp.MustCompile("^[\x00-\x7f]+$")

// Validate checks the field values on BytesPrefix with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesPrefix) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesPrefix with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesPrefixMultiError, or
// nil if none found.
func (m *BytesPrefix) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesPrefix) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !bytes.HasPrefix(m.GetVal(), []uint8{0x99}) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not have prefix \"\\x99\"",
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

// Validate checks the field values on BytesContains with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesContains) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesContains with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesContainsMultiError, or
// nil if none found.
func (m *BytesContains) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesContains) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !bytes.Contains(m.GetVal(), []uint8{0x62, 0x61, 0x72}) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not contain \"\\x62\\x61\\x72\"",
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

// Validate checks the field values on BytesSuffix with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesSuffix) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesSuffix with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesSuffixMultiError, or
// nil if none found.
func (m *BytesSuffix) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesSuffix) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if !bytes.HasSuffix(m.GetVal(), []uint8{0x62, 0x75, 0x7A, 0x7A}) {
		err := ValidationError{
			field:   "Val",
			reason:  "value does not have suffix \"\\x62\\x75\\x7A\\x7A\"",
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

// Validate checks the field values on BytesIP with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesIP) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesIP with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in BytesIPMultiError, or nil if none found.
func (m *BytesIP) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesIP) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.IP(m.GetVal()); ip.To16() == nil {
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

// Validate checks the field values on BytesIPv4 with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesIPv4) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesIPv4 with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesIPv4MultiError, or nil
// if none found.
func (m *BytesIPv4) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesIPv4) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.IP(m.GetVal()); ip.To4() == nil {
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

// Validate checks the field values on BytesIPv6 with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BytesIPv6) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesIPv6 with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BytesIPv6MultiError, or nil
// if none found.
func (m *BytesIPv6) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesIPv6) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if ip := net.IP(m.GetVal()); ip.To16() == nil || ip.To4() != nil {
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

// Validate checks the field values on BytesIPv6Ignore with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *BytesIPv6Ignore) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BytesIPv6Ignore with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// BytesIPv6IgnoreMultiError, or nil if none found.
func (m *BytesIPv6Ignore) ValidateAll() error {
	return m.validate(true)
}

func (m *BytesIPv6Ignore) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetVal()) > 0 {

		if ip := net.IP(m.GetVal()); ip.To16() == nil || ip.To4() != nil {
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

	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}
