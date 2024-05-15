// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v4.25.0--rc2
// source: cases/wkt_any.proto

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

// Validate checks the field values on AnyRequired with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *AnyRequired) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AnyRequired with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in AnyRequiredMultiError, or
// nil if none found.
func (m *AnyRequired) ValidateAll() error {
	return m.validate(true)
}

func (m *AnyRequired) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetVal() == nil {
		err := ValidationError{
			field:   "Val",
			reason:  "value is required",
			errCode: 0,
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if a := m.GetVal(); a != nil {

	}

	if len(errors) > 0 {
		return MultiError(errors)
	}

	return nil
}

// Validate checks the field values on AnyIn with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *AnyIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AnyIn with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in AnyInMultiError, or nil if none found.
func (m *AnyIn) ValidateAll() error {
	return m.validate(true)
}

func (m *AnyIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if a := m.GetVal(); a != nil {

		if _, ok := _AnyIn_Val_InLookup[a.GetTypeUrl()]; !ok {
			err := ValidationError{
				field:   "Val",
				reason:  "type URL must be in list [type.googleapis.com/google.protobuf.Duration]",
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

var _AnyIn_Val_InLookup = map[string]struct{}{
	"type.googleapis.com/google.protobuf.Duration": {},
}

// Validate checks the field values on AnyNotIn with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *AnyNotIn) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on AnyNotIn with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in AnyNotInMultiError, or nil
// if none found.
func (m *AnyNotIn) ValidateAll() error {
	return m.validate(true)
}

func (m *AnyNotIn) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if a := m.GetVal(); a != nil {

		if _, ok := _AnyNotIn_Val_NotInLookup[a.GetTypeUrl()]; ok {
			err := ValidationError{
				field:   "Val",
				reason:  "type URL must not be in list [type.googleapis.com/google.protobuf.Timestamp]",
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

var _AnyNotIn_Val_NotInLookup = map[string]struct{}{
	"type.googleapis.com/google.protobuf.Timestamp": {},
}
