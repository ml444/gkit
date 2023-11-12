package httpmw

import (
	"reflect"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

func CheckResponseError() middleware.AfterHandler {
	return func(rsp interface{}, err error) (interface{}, error) {
		if err != nil {
			if _, ok := err.(*errorx.Error); !ok {
				err = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeUnknown, err.Error())
			}
		}
		return rsp, err
	}
}

// ReplaceEmptyResponse replace empty response with specified information.
func ReplaceEmptyResponse(data interface{}) middleware.AfterHandler {
	return func(rsp interface{}, err error) (interface{}, error) {
		if err == nil && (rsp == nil || reflect.ValueOf(rsp).Elem().IsZero()) {
			return data, nil
		}
		return rsp, err
	}
}
