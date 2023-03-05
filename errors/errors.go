package errors

import (
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

const (
	// UnknownCode is unknown code for error info.
	UnknownCode = 500
	// UnknownErrCode is unknown errCode for error info.
	UnknownErrCode = -1
)

var errCodeMap = map[int32]string{}

func RegisterError(m map[int32]string) {
	for k, v := range m {
		errCodeMap[k] = v
	}
}

type Status struct {
	StatusCode int32
	Message    string
	Metadata   map[string]string
}

// Error is a status error.
type Error struct {
	Status
	ErrCode int32
	cause   error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d errcode = %d message = '%s' metadata = %v cause = %s", e.StatusCode, e.ErrCode, e.Message, e.Metadata, e.cause)
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *Error) Unwrap() error { return e.cause }

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.StatusCode == e.StatusCode && se.ErrCode == e.ErrCode
	}
	return false
}

func (e *Error) GetStatusCode() int32 {
	return e.Status.StatusCode
}

func (e *Error) GetErrCode() int32 {
	return e.ErrCode
}

// WithCause with the underlying cause of the error.
func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

// WithMetadata with an MD formed by the mapping of key, value.
func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md
	return err
}

// GRPCStatus returns the Status represented by se.
func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(ToGRPCCode(int(e.StatusCode)), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Message,
			Metadata: e.Metadata,
		})
	return s
}

func New(code int, errCode int32) *Error {
	msg := errCodeMap[errCode]
	return &Error{
		Status: Status{
			StatusCode: int32(code),
			Message:    msg,
		},
		ErrCode: errCode,
	}
}

// NewWithMsg returns an error object for the code, message.
func NewWithMsg(code int, errCode int32, message string) *Error {
	return &Error{
		Status: Status{
			StatusCode: int32(code),
			Message:    message,
		},
		ErrCode: errCode,
	}
}

// NewWithMsgf NewWithMsg(code fmt.Sprintf(format, a...))
func NewWithMsgf(code int, errCode int32, format string, a ...interface{}) *Error {
	return NewWithMsg(code, errCode, fmt.Sprintf(format, a...))
}

// Errorf returns an error object for the code, message and error info.
func Errorf(code int, errCode int32, format string, a ...interface{}) error {
	return NewWithMsg(code, errCode, fmt.Sprintf(format, a...))
}

// Code returns the http code for an error.
// It supports wrapped errors.
func Code(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}
	return int(FromError(err).StatusCode)
}

// ErrCode returns the reason for a particular error.
// It supports wrapped errors.
func ErrCode(err error) int32 {
	if err == nil {
		return UnknownErrCode
	}
	return FromError(err).ErrCode
}

// Clone deep clone error to a new error.
func Clone(err *Error) *Error {
	if err == nil {
		return nil
	}
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		cause: err.cause,
		Status: Status{
			StatusCode: err.StatusCode,
			Message:    err.Message,
			Metadata:   metadata,
		},
		ErrCode: err.ErrCode,
	}
}

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if !ok {
		return NewWithMsg(UnknownCode, UnknownErrCode, err.Error())
	}
	ret := NewWithMsg(
		FromGRPCCode(gs.Code()),
		UnknownErrCode,
		gs.Message(),
	)
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Message = d.Reason
			return ret.WithMetadata(d.Metadata)
		}
	}
	return ret
}
