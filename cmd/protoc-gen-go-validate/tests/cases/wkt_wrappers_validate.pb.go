// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v3.21.5
// source: cases/wkt_wrappers.proto

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
	//"google.golang.org/protobuf/types/known/anypb"
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

// Validate checks the field values on WrapperFloat with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperFloat) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperFloat with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperFloatMultiError, or
// nil if none found.
func (m *WrapperFloat) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperFloat) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperDouble with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperDouble) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperDouble with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperDoubleMultiError, or
// nil if none found.
func (m *WrapperDouble) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperDouble) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperInt64 with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperInt64) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperInt64 with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperInt64MultiError, or
// nil if none found.
func (m *WrapperInt64) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperInt64) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperInt32 with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperInt32) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperInt32 with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperInt32MultiError, or
// nil if none found.
func (m *WrapperInt32) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperInt32) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperUInt64 with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperUInt64) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperUInt64 with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperUInt64MultiError, or
// nil if none found.
func (m *WrapperUInt64) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperUInt64) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperUInt32 with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperUInt32) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperUInt32 with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperUInt32MultiError, or
// nil if none found.
func (m *WrapperUInt32) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperUInt32) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
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

// Validate checks the field values on WrapperBool with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperBool) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperBool with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperBoolMultiError, or
// nil if none found.
func (m *WrapperBool) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperBool) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() != true {
			err := ValidationError{
				field:  "Val",
				reason: "value must equal true",
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

// Validate checks the field values on WrapperString with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperString) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperString with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperStringMultiError, or
// nil if none found.
func (m *WrapperString) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperString) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if !strings.HasSuffix(wrapper.GetValue(), "bar") {
			err := ValidationError{
				field:  "Val",
				reason: "value does not have suffix \"bar\"",
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

// Validate checks the field values on WrapperBytes with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WrapperBytes) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperBytes with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WrapperBytesMultiError, or
// nil if none found.
func (m *WrapperBytes) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperBytes) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if len(wrapper.GetValue()) < 3 {
			err := ValidationError{
				field:  "Val",
				reason: "value length must be at least 3 bytes",
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

// Validate checks the field values on WrapperRequiredString with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *WrapperRequiredString) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperRequiredString with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// WrapperRequiredStringMultiError, or nil if none found.
func (m *WrapperRequiredString) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperRequiredString) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() != "bar" {
			err := ValidationError{
				field:  "Val",
				reason: "value must equal bar",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	} else {
		err := ValidationError{
			field:  "Val",
			reason: "value is required and must not be nil.",
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

// Validate checks the field values on WrapperRequiredEmptyString with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *WrapperRequiredEmptyString) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperRequiredEmptyString with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// WrapperRequiredEmptyStringMultiError, or nil if none found.
func (m *WrapperRequiredEmptyString) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperRequiredEmptyString) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() != "" {
			err := ValidationError{
				field:  "Val",
				reason: "value must equal ",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	} else {
		err := ValidationError{
			field:  "Val",
			reason: "value is required and must not be nil.",
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

// Validate checks the field values on WrapperOptionalUuidString with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *WrapperOptionalUuidString) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperOptionalUuidString with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// WrapperOptionalUuidStringMultiError, or nil if none found.
func (m *WrapperOptionalUuidString) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperOptionalUuidString) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if err := _validateUuid(wrapper.GetValue()); err != nil {
			err = ValidationError{
				field:  "Val",
				reason: "value must be a valid UUID",
				cause:  err,
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

// Validate checks the field values on WrapperRequiredFloat with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *WrapperRequiredFloat) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WrapperRequiredFloat with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// WrapperRequiredFloatMultiError, or nil if none found.
func (m *WrapperRequiredFloat) ValidateAll() error {
	return m.validate(true)
}

func (m *WrapperRequiredFloat) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if wrapper := m.GetVal(); wrapper != nil {

		if wrapper.GetValue() <= 0 {
			err := ValidationError{
				field:  "Val",
				reason: "value must be greater than 0",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	} else {
		err := ValidationError{
			field:  "Val",
			reason: "value is required and must not be nil.",
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
