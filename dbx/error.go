package dbx

import (
	"errors"
	"net/http"
	"strings"

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
		eCode := Err.GetCode()
		if eCode == errorx.ErrCodeRecordNotFoundSys || eCode == errCode {
			return true
		}
	}
	if gs, ok := status.FromError(err); ok && gs.Code() == codes.NotFound {
		return true
	}
	return false
}

var ErrUpdateRowAffectedZero = errorx.CreateError(400, errorx.ErrCodeUpdateRowAffectedZeroSys, "update row affected zero") //nolint:gochecknoglobals

func IsUpdateRowAffectedZero(err error) bool {
	var Err *errorx.Error
	if errors.As(err, &Err) {
		return Err.GetCode() == errorx.ErrCodeUpdateRowAffectedZeroSys
	}
	return false
}

const mysqlDuplicateEntryCode uint16 = 1062

func IsDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	// var mysqlErr *mysqlDriver.MySQLError
	// if errors.As(err, &mysqlErr) {
	// 	return mysqlErr.Number == mysqlDuplicateEntryCode
	// }
	msg := err.Error()
	if strings.Contains(msg, "Duplicate entry") || // MySQL
		strings.Contains(msg, "duplicate key") || // PostgreSQL
		strings.Contains(msg, "UNIQUE constraint failed") { // SQLite
		return true
	}
	return false
}
