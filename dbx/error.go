package dbx

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ml444/gkit/errorx"
)

func GetNotFoundErr(err error) *errorx.Error {
	return errorx.CreateError(
		http.StatusNotFound,
		errorx.ErrCodeRecordNotFoundSys,
		err.Error(),
	)
}

// IsNotFoundErr returns true if err is an *Error and its ErrorCode matches ErrCodeRecordNotFoundSys or errCode.
func IsNotFoundErr(err error, errCode int32) bool {
	var Err *errorx.Error
	if errors.As(err, &Err) {
		eCode := Err.GetErrorCode()
		if eCode == errorx.ErrCodeRecordNotFoundSys || eCode == errCode {
			return true
		}
	}
	if gs, ok := status.FromError(err); ok && gs.Code() == codes.NotFound {
		return true
	}
	return false
}
