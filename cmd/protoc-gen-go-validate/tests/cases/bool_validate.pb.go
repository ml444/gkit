// Code generated by protoc-gen-go-validate. DO NOT EDIT.
// - protoc-gen-go-validate 1.0.0
// - protoc             v4.25.0--rc2
// source: cases/bool.proto

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

// Validate checks the field values on BoolConstTrue with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BoolConstTrue) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BoolConstTrue with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BoolConstTrueMultiError, or
// nil if none found.
func (m *BoolConstTrue) ValidateAll() error {
	return m.validate(true)
}

func (m *BoolConstTrue) validate(all bool) error {
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

// Validate checks the field values on BoolConstFalse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *BoolConstFalse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on BoolConstFalse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in BoolConstFalseMultiError,
// or nil if none found.
func (m *BoolConstFalse) ValidateAll() error {
	return m.validate(true)
}

func (m *BoolConstFalse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetVal() != false {
		err := ValidationError{
			field:   "Val",
			reason:  "value must equal false",
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
