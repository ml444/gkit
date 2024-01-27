// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v3.21.5
// source: cases/wkt_nested.proto

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

// Validate checks the field values on WktLevelOne with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WktLevelOne) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WktLevelOne with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in WktLevelOneMultiError, or
// nil if none found.
func (m *WktLevelOne) ValidateAll() error {
	return m.validate(true)
}

func (m *WktLevelOne) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetTwo() == nil {
		err := ValidationError{
			field:  "Two",
			reason: "value is required",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetTwo()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ValidationError{
					field:  "Two",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ValidationError{
					field:  "Two",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetTwo()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ValidationError{
				field:  "Two",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on WktLevelOne_WktLevelTwo with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *WktLevelOne_WktLevelTwo) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WktLevelOne_WktLevelTwo with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// WktLevelOne_WktLevelTwoMultiError, or nil if none found.
func (m *WktLevelOne_WktLevelTwo) ValidateAll() error {
	return m.validate(true)
}

func (m *WktLevelOne_WktLevelTwo) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetThree() == nil {
		err := ValidationError{
			field:  "Three",
			reason: "value is required",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetThree()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ValidationError{
					field:  "Three",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ValidationError{
					field:  "Three",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetThree()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ValidationError{
				field:  "Three",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on WktLevelOne_WktLevelTwo_WktLevelThree
// with the rules defined in the proto definition for this message. If any
// rules are violated, the first error encountered is returned, or nil if
// there are no violations.
func (m *WktLevelOne_WktLevelTwo_WktLevelThree) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WktLevelOne_WktLevelTwo_WktLevelThree
// with the rules defined in the proto definition for this message. If any
// rules are violated, the result is a list of violation errors wrapped in
// WktLevelOne_WktLevelTwo_WktLevelThreeMultiError, or nil if none found.
func (m *WktLevelOne_WktLevelTwo_WktLevelThree) ValidateAll() error {
	return m.validate(true)
}

func (m *WktLevelOne_WktLevelTwo_WktLevelThree) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if err := _validateUuid(m.GetUuid()); err != nil {
		err = ValidationError{
			field:  "Uuid",
			reason: "value must be a valid UUID",
			cause:  err,
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