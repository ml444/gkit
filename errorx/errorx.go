package errorx

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrCodeDetail struct {
	StatusCode int32
	ErrorCode  int32
	Message    string
	Polyglot   map[string]string
}

var lang string
var errCodeMap = map[int32]*ErrCodeDetail{}
var defaultError = &ErrCodeDetail{}

func SetLang(l string) {
	lang = l
}

func RegisterError(codeMap map[int32]*ErrCodeDetail, defaultErr *ErrCodeDetail) {
	for k, detail := range codeMap {
		if detail.StatusCode == 0 {
			detail.StatusCode = DefaultStatusCode
		}
		errCodeMap[k] = detail
	}
	if defaultErr.StatusCode == 0 {
		defaultErr.StatusCode = DefaultStatusCode
	}
	if defaultError != nil {
		defaultError = defaultErr
	}
}

// Error is a status error.
type Error struct {
	ErrorInfo
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code=%d errcode=%d message='%s' metadata=%v cause=%s", e.StatusCode, e.ErrorCode, e.Message, e.Metadata, e.cause)
}

// JSONBytes returns the JSON representation of the error.
func (e *Error) JSONBytes() []byte {
	buf, _ := json.Marshal(e)
	return buf
}

func (e *Error) Unwrap() error { return e.cause }

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

func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)
	err.cause = cause
	return err
}

func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)
	err.Metadata = md
	return err
}

func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(ToGRPCCode(int(e.StatusCode)), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Message,
			Metadata: e.Metadata,
		})
	return s
}

func pickMsg(detail *ErrCodeDetail) string {
	msg := detail.Message
	if detail.Polyglot != nil && lang != "" {
		if v, ok := detail.Polyglot[lang]; ok {
			msg = v
		}
	}
	return msg
}

func getErrDetail(errCode int32) *ErrCodeDetail {
	detail, ok := errCodeMap[errCode]
	if !ok {
		detail = &ErrCodeDetail{}
		detail.StatusCode = UnknownStatusCode
	}
	return detail
}

func New(errCode int32) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			ErrorCode:  errCode,
			StatusCode: detail.StatusCode,
			Message:    pickMsg(detail),
		},
	}
}

// NewWithStatus new error from errcode and message
func NewWithStatus(statusCode int32, errCode int32) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			ErrorCode:  errCode,
			StatusCode: statusCode,
			Message:    pickMsg(detail),
		},
	}
}

// NewWithMsg new error from errcode and message
func NewWithMsg(errCode int32, msg string) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			ErrorCode:  errCode,
			StatusCode: detail.StatusCode,
			Message:    msg,
		},
	}
}

func NewDefaultError(statusCode int32, message string) *Error {
	if statusCode == 0 {
		statusCode = DefaultStatusCode
	}
	if message == "" {
		message = defaultError.Message
	}
	return CreateError(statusCode, defaultError.ErrorCode, message)
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

func CreateErrorf(statusCode int32, errCode int32, format string, a ...interface{}) *Error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

func Errorf(statusCode int32, errCode int32, format string, a ...interface{}) error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

// StatusCode returns the http code for an error.
// It supports wrapped errorx.
func StatusCode(err error) int {
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
	var Err *Error
	if errors.As(err, &Err) {
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
