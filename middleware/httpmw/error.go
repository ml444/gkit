package httpmw

import (
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

func ConvertResponseError() middleware.AfterHandler {
	return func(rsp interface{}, err error) (interface{}, error) {
		if err != nil {
			return rsp, errorx.FromError(err)
		}
		return rsp, err
	}
}
