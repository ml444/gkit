package jwt

import (
	"github.com/ml444/gkit/errorx"
	"net/http"
)

var ErrTokenFormat = func() error {
	return errorx.CreateError(
		http.StatusForbidden,
		errorx.ErrCodeInvalidHeaderSys,
		"the Authorization token is incorrectly formatted",
	)
}

var ErrClaims = func() error {
	return errorx.CreateError(
		http.StatusUnauthorized,
		errorx.ErrCodeInvalidHeaderSys,
		"Claims assertion failure",
	)
}
