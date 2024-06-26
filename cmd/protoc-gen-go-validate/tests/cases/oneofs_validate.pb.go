// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v4.25.0--rc2
// source: cases/oneofs.proto

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

// Validate checks the field values on TestOneOfMsg with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *TestOneOfMsg) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestOneOfMsg with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in TestOneOfMsgMultiError, or
// nil if none found.
func (m *TestOneOfMsg) ValidateAll() error {
	return m.validate(true)
}

func (m *TestOneOfMsg) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetVal() != true {
		err := ValidationError{
			field:   "Val",
			reason:  "value must equal true",
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

// Validate checks the field values on OneOf with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *OneOf) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on OneOf with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in OneOfMultiError, or nil if none found.
func (m *OneOf) ValidateAll() error {
	return m.validate(true)
}

func (m *OneOf) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	switch v := m.O.(type) {
	case *OneOf_X:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if !strings.HasPrefix(m.GetX(), "foo") {
			err := ValidationError{
				field:   "X",
				reason:  "value does not have prefix \"foo\"",
				errCode: 0,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	case *OneOf_Y:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if m.GetY() <= 0 {
			err := ValidationError{
				field:   "Y",
				reason:  "value must be greater than 0",
				errCode: 0,
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	case *OneOf_Z:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}
		// no validation rules for Z
	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on OneOfIgnoreEmpty with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *OneOfIgnoreEmpty) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on OneOfIgnoreEmpty with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// OneOfIgnoreEmptyMultiError, or nil if none found.
func (m *OneOfIgnoreEmpty) ValidateAll() error {
	return m.validate(true)
}

func (m *OneOfIgnoreEmpty) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	switch v := m.O.(type) {
	case *OneOfIgnoreEmpty_X:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if m.GetX() != "" {

			if l := utf8.RuneCountInString(m.GetX()); l < 3 || l > 5 {
				err := ValidationError{
					field:   "X",
					reason:  "value length must be between 3 and 5 runes, inclusive",
					errCode: 0,
				}
				if !all {
					return err
				}
				errors = append(errors, err)
			}

		}

	case *OneOfIgnoreEmpty_Y:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if len(m.GetY()) > 0 {

			if l := len(m.GetY()); l < 3 || l > 5 {
				err := ValidationError{
					field:   "Y",
					reason:  "value length must be between 3 and 5 bytes, inclusive",
					errCode: 0,
				}
				if !all {
					return err
				}
				errors = append(errors, err)
			}

		}

	case *OneOfIgnoreEmpty_Z:
		if v == nil {
			err := ValidationError{
				field:  "O",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if m.GetZ() != 0 {

			if val := m.GetZ(); val > 128 && val < 256 {
				err := ValidationError{
					field:   "Z",
					reason:  "value must be outside range (128, 256)",
					errCode: 0,
				}
				if !all {
					return err
				}
				errors = append(errors, err)
			}

		}

	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}
