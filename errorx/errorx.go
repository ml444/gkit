// Package errorx provides structured errors with HTTP status, business error codes,
// optional metadata, i18n messages, and gRPC status conversion.
//
// Error embeds the protobuf ErrorInfo message for JSON/gRPC wire compatibility.
// A separate domain struct is intentionally not used to avoid duplicate fields and
// keep httpx codec marshaling aligned with the proto schema.
package errorx

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrCodeDetail struct {
	Status   int32
	Code     int32
	Message  string
	Polyglot map[string]string
}

var (
	gMsgLanguage atomic.Value // string
	errCodeMap   = map[int32]*ErrCodeDetail{}
	lock         sync.RWMutex
)

func SetLang(l string) {
	gMsgLanguage.Store(l)
}

func msgLanguage() string {
	v, _ := gMsgLanguage.Load().(string)
	return v
}

func cloneErrCodeDetail(d *ErrCodeDetail) *ErrCodeDetail {
	if d == nil {
		return nil
	}
	cp := &ErrCodeDetail{
		Status:  d.Status,
		Code:    d.Code,
		Message: d.Message,
	}
	if len(d.Polyglot) > 0 {
		cp.Polyglot = make(map[string]string, len(d.Polyglot))
		for k, v := range d.Polyglot {
			cp.Polyglot[k] = v
		}
	}
	return cp
}

func RegisterError(codeMap map[int32]*ErrCodeDetail) {
	lock.Lock()
	defer lock.Unlock()
	for k, detail := range codeMap {
		cp := cloneErrCodeDetail(detail)
		if cp.Status == 0 {
			cp.Status = DefaultStatusCode
		}
		errCodeMap[k] = cp
	}
}

// Error is a status error.
type Error struct {
	ErrorInfo
	cause error
}

func (e *Error) Error() string {
	if e.cause != nil && len(e.Metadata) != 0 {
		return fmt.Sprintf("error: [%d:%d] '%s' metadata=%v cause=%s", e.Status, e.Code, e.Message, e.Metadata, e.cause)
	} else if e.cause != nil {
		return fmt.Sprintf("error: [%d:%d] '%s' cause=%s", e.Status, e.Code, e.Message, e.cause)
	} else if len(e.Metadata) != 0 {
		return fmt.Sprintf("error: [%d:%d] '%s' metadata=%v", e.Status, e.Code, e.Message, e.Metadata)
	} else {
		return fmt.Sprintf("error: [%d:%d] '%s'", e.Status, e.Code, e.Message)
	}
}

// JSONBytes returns the JSON representation of the error.
func (e *Error) JSONBytes() []byte {
	buf, _ := json.Marshal(e)
	return buf
}

func (e *Error) Unwrap() error { return e.cause }

// Is reports whether err is an *Error with the same HTTP status and business code.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Status == e.Status && se.Code == e.Code
	}
	return false
}

func (e *Error) IsErrCode(errCode int32) bool {
	return e.Code == errCode
}

func (e *Error) GetStatus() int32 {
	return e.ErrorInfo.Status
}

func (e *Error) GetCode() int32 {
	return e.ErrorInfo.Code
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

// ConvertMsgByLang picks the first Accept-Language tag that has a registered translation.
// When a requested tag matches the server default language (SetLang), the message from
// New()/pickMsg is kept unchanged.
func (e *Error) ConvertMsgByLang(langs ...string) {
	if len(langs) == 0 {
		return
	}
	lock.RLock()
	detail, ok := errCodeMap[e.Code]
	lock.RUnlock()
	if !ok || len(detail.Polyglot) == 0 {
		return
	}
	defaultLang := msgLanguage()
	for _, lang := range langs {
		if defaultLang == lang {
			return
		}
		msg, ok := detail.Polyglot[lang]
		if ok {
			e.Message = msg
			return
		}
	}
}

func (e *Error) GRPCStatus() *status.Status {
	st := status.New(ToGRPCCode(int(e.Status)), e.Message)
	s, err := st.WithDetails(&ErrorInfo{
		Status:   e.Status,
		Code:     e.Code,
		Message:  e.Message,
		Metadata: e.Metadata,
	})
	if err != nil {
		return st
	}
	return s
}

func pickMsg(detail *ErrCodeDetail) string {
	msg := detail.Message
	if detail.Polyglot != nil {
		if v, ok := detail.Polyglot[msgLanguage()]; ok {
			msg = v
		}
	}
	return msg
}

func getErrDetail(errCode int32) *ErrCodeDetail {
	lock.RLock()
	detail, ok := errCodeMap[errCode]
	lock.RUnlock()
	if !ok {
		return &ErrCodeDetail{Status: DefaultStatusCode}
	}
	return detail
}

func New(errCode int32) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			Code:    errCode,
			Status:  detail.Status,
			Message: pickMsg(detail),
		},
	}
}

// NewWithStatus new error from errcode and message
func NewWithStatus(statusCode int32, errCode int32) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			Code:    errCode,
			Status:  statusCode,
			Message: pickMsg(detail),
		},
	}
}

// NewWithMsg new error from errcode and message
func NewWithMsg(errCode int32, msg string) *Error {
	detail := getErrDetail(errCode)
	return &Error{
		ErrorInfo: ErrorInfo{
			Code:    errCode,
			Status:  detail.Status,
			Message: msg,
		},
	}
}

// CreateError returns an error object for the status code, error code, message.
func CreateError(statusCode int32, errCode int32, message string) *Error {
	return &Error{
		ErrorInfo: ErrorInfo{
			Status:  statusCode,
			Code:    errCode,
			Message: message,
		},
	}
}

func CreateErrorf(statusCode int32, errCode int32, format string, a ...interface{}) *Error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

func Errorf(statusCode int32, errCode int32, format string, a ...interface{}) error {
	return CreateError(statusCode, errCode, fmt.Sprintf(format, a...))
}

// Status returns the http code for an error.
// It supports wrapped errorx.
func Status(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}
	return int(FromError(err).Status)
}

// ErrCode returns the reason for a particular error.
// It supports wrapped errorx.
func ErrCode(err error) int32 {
	if err == nil {
		return ErrCodeUnknown
	}
	return FromError(err).Code
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
			Status:   err.Status,
			Code:     err.Code,
			Message:  err.Message,
			Metadata: metadata,
		},
	}
}

// FromError try to convert an error to *Error.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if e := new(Error); errors.As(err, &e) {
		return e
	}
	gs, ok := status.FromError(err)
	if !ok {
		return CreateError(UnknownStatusCode, ErrCodeUnknown, err.Error()).WithCause(err)
	}
	ret := CreateError(
		int32(FromGRPCCode(gs.Code())),
		ErrCodeUnknown,
		gs.Message(),
	).WithCause(err)
	for _, detail := range gs.Details() {
		if errInfo, ok := detail.(*ErrorInfo); ok {
			out := CreateError(
				errInfo.Status,
				errInfo.Code,
				errInfo.Message,
			).WithCause(err)
			if len(errInfo.Metadata) > 0 {
				out = out.WithMetadata(errInfo.Metadata)
			}
			return out
		}
		if legacy, ok := detail.(*errdetails.ErrorInfo); ok {
			out := ret
			if md := legacy.GetMetadata(); len(md) > 0 {
				out = out.WithMetadata(md)
			}
			return out
		}
	}
	return ret
}

// ErrorIs returns true if err is an *Error and its Code matches errCode.
func ErrorIs(err error, errCode int32) bool {
	var e *Error
	ok := errors.As(err, &e)
	if ok {
		eCode := e.GetCode()
		if eCode == errCode {
			return true
		}
	}
	return false
}
