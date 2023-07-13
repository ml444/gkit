package errorx

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

const DefaultErrorCodeLowerLimit = 100000

type ErrCodeDetail struct {
	StatusCode int32
	Message    string
}

var errCodeMap = map[int32]*ErrCodeDetail{}

func RegisterError(msgMap map[int32]string, codeMap map[int32]int32) {
	for k, v := range msgMap {
		detail := ErrCodeDetail{}
		detail.Message = v
		if statusCode, ok := codeMap[k]; ok {
			detail.StatusCode = statusCode
		} else {
			if k > DefaultErrorCodeLowerLimit {
				detail.StatusCode = DefaultStatusCode
			}
		}
		errCodeMap[k] = &detail
	}
}

// Error is a status error.
type Error struct {
	ErrorInfo
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d errcode = %d message = '%s' metadata = %v cause = %s", e.StatusCode, e.ErrorCode, e.Message, e.Metadata, e.cause)
}

// JSONBytes returns the JSON representation of the error.
func (e *Error) JSONBytes() []byte {
	buf, _ := json.Marshal(e)
	return buf
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *Error) Unwrap() error { return e.cause }

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.StatusCode == e.StatusCode && se.ErrorCode == e.ErrorCode
	}
	return false
}
func (e *Error) IsErrCode(errCode int32) bool {
	return e.ErrorCode == errCode
}

func (e *Error) GetStatusCode() int32 {
	return e.ErrorInfo.StatusCode
}

func (e *Error) GetErrorCode() int32 {
	return e.ErrorInfo.ErrorCode
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

func New(errCode int32) *Error {
	detail, ok := errCodeMap[errCode]
	if !ok {
		detail = &ErrCodeDetail{}
		detail.StatusCode = UnknownStatusCode
		detail.Message = ""
	}
	return &Error{
		ErrorInfo: ErrorInfo{
			ErrorCode:  errCode,
			StatusCode: detail.StatusCode,
			Message:    detail.Message,
		},
	}
}

// NewWithMsg new error from errcode and message
func NewWithMsg(errCode int32, msg string) *Error {
	detail, ok := errCodeMap[errCode]
	if !ok {
		detail = &ErrCodeDetail{}
		detail.StatusCode = UnknownStatusCode
		detail.Message = msg
	}
	return &Error{
		ErrorInfo: ErrorInfo{
			ErrorCode:  errCode,
			StatusCode: detail.StatusCode,
			Message:    detail.Message,
		},
	}
}

// CreateError returns an error object for the status code, error code, message.
func CreateError(statusCode int32, errCode int32, message string) *Error {
	return &Error{
		ErrorInfo: ErrorInfo{
			StatusCode: statusCode,
			ErrorCode:  errCode,
			Message:    message,
		},
	}
}

// CreateErrorf CreateError(code fmt.Sprintf(format, a...))
func CreateErrorf(statusCode int32, errCode int32, format string, a ...interface{}) *Error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

// Errorf CreateError(code fmt.Sprintf(format, a...))
func Errorf(statusCode int32, errCode int32, format string, a ...interface{}) error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

// Code returns the http code for an error.
// It supports wrapped errorx.
func Code(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}
	return int(FromError(err).StatusCode)
}

// ErrCode returns the reason for a particular error.
// It supports wrapped errorx.
func ErrCode(err error) int32 {
	if err == nil {
		return ErrCodeUnknown
	}
	return FromError(err).ErrorCode
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
		ErrorInfo: ErrorInfo{
			StatusCode: err.StatusCode,
			ErrorCode:  err.ErrorCode,
			Message:    err.Message,
			Metadata:   metadata,
		},
	}
}

// FromError try to convert an error to *Error.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if Err, ok := err.(*Error); ok {
		return Err
	}
	// wrapped error
	//if se := new(Error); errors.As(err, &se) {
	//	return se
	//}
	gs, ok := status.FromError(err)
	if !ok {
		return CreateError(UnknownStatusCode, ErrCodeUnknown, err.Error())
	}
	ret := CreateError(
		int32(FromGRPCCode(gs.Code())),
		ErrCodeUnknown,
		gs.Message(),
	)
	for _, detail := range gs.Details() {
		if ErrInfo, ok := detail.(*ErrorInfo); ok {
			return CreateError(
				ErrInfo.StatusCode,
				ErrInfo.ErrorCode,
				ErrInfo.Message,
			).WithMetadata(ErrInfo.Metadata)
		}
	}
	return ret
}

// ErrorAs returns true if err is an *Error and its ErrorCode matches errCode.
func ErrorAs(err error, errCode int32) bool {
	e, ok := err.(*Error)
	if ok {
		eCode := e.GetErrorCode()
		if eCode == errCode {
			return true
		}
	}
	return false
}

// IsNotFoundErr returns true if err is an *Error and its ErrorCode matches ErrCodeRecordNotFoundSys or errCode.
func IsNotFoundErr(err error, errCode int32) bool {
	e, ok := err.(*Error)
	if ok {
		eCode := e.GetErrorCode()
		if eCode == ErrCodeRecordNotFoundSys || eCode == errCode {
			return true
		}
	}
	return false
}
