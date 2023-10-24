// nolint:gomnd
package errorx

import "net/http"

// BadRequest new BadRequest error that is mapped to a 400 response.
func BadRequest(message string) *Error {
	return CreateError(http.StatusBadRequest, ErrCodeInvalidReqSys, message)
}

// IsBadRequest determines if err is an error which indicates a BadRequest error.
// It supports wrapped errors.
func IsBadRequest(err error) bool {
	return StatusCode(err) == http.StatusBadRequest
}

// Unauthorized new Unauthorized error that is mapped to a 401 response.
func Unauthorized(message string) *Error {
	return CreateError(http.StatusUnauthorized, ErrCodeInvalidHeaderSys, message)
}

// IsUnauthorized determines if err is an error which indicates an Unauthorized error.
// It supports wrapped errors.
func IsUnauthorized(err error) bool {
	return StatusCode(err) == http.StatusUnauthorized
}

// Forbidden new Forbidden error that is mapped to a 403 response.
func Forbidden(message string) *Error {
	return CreateError(http.StatusForbidden, ErrCodeInvalidBodySys, message)
}

// IsForbidden determines if err is an error which indicates a Forbidden error.
// It supports wrapped errors.
func IsForbidden(err error) bool {
	return StatusCode(err) == http.StatusForbidden
}

// NotFound new NotFound error that is mapped to a 404 response.
func NotFound(message string) *Error {
	return CreateError(http.StatusNotFound, ErrCodeRecordNotFoundSys, message)
}

// IsNotFound determines if err is an error which indicates an NotFound error.
// It supports wrapped errors.
func IsNotFound(err error) bool {
	return StatusCode(err) == http.StatusNotFound
}

// Conflict new Conflict error that is mapped to a 409 response.
func Conflict(message string) *Error {
	return CreateError(http.StatusConflict, ErrCodeUnknown, message)
}

// IsConflict determines if err is an error which indicates a Conflict error.
// It supports wrapped errors.
func IsConflict(err error) bool {
	return StatusCode(err) == http.StatusConflict
}

// InternalServer new InternalServer error that is mapped to a 500 response.
func InternalServer(message string) *Error {
	return CreateError(http.StatusInternalServerError, ErrCodeUnknown, message)
}

// IsInternalServer determines if err is an error which indicates an Internal error.
// It supports wrapped errors.
func IsInternalServer(err error) bool {
	return StatusCode(err) == http.StatusInternalServerError
}

// ServiceUnavailable new ServiceUnavailable error that is mapped to an HTTP 503 response.
func ServiceUnavailable(message string) *Error {
	return CreateError(http.StatusInternalServerError, ErrCodeUnknown, message)
}

// IsServiceUnavailable determines if err is an error which indicates an Unavailable error.
// It supports wrapped errors.
func IsServiceUnavailable(err error) bool {
	return StatusCode(err) == http.StatusInternalServerError
}

// GatewayTimeout new GatewayTimeout error that is mapped to an HTTP 504 response.
func GatewayTimeout(message string) *Error {
	return CreateError(http.StatusInternalServerError, ErrCodeUnknown, message)
}

// IsGatewayTimeout determines if err is an error which indicates a GatewayTimeout error.
// It supports wrapped errors.
func IsGatewayTimeout(err error) bool {
	return StatusCode(err) == 504
}

// ClientClosed new ClientClosed error that is mapped to an HTTP 499 response.
func ClientClosed(message string) *Error {
	return CreateError(StatusClientClosed, ErrCodeUnknown, message)
}

// IsClientClosed determines if err is an error which indicates a IsClientClosed error.
// It supports wrapped errors.
func IsClientClosed(err error) bool {
	return StatusCode(err) == 499
}
