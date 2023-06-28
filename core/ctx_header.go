package core

import (
	"context"
	"net/http"

	"github.com/ml444/gkit/errorx"
)

const (
	HeadersKey    = "headers"
	UserIdKey     = "user_id"
	CorpIdKey     = "corp_id"
	ClientTypeKey = "client_type"
	ClientIdKey   = "client_id"
	RemoteIp      = "remote_ip"
)

type Header map[string]interface{}

var NotFoundHeaders = errorx.CreateError(
	http.StatusForbidden,
	errorx.ErrCodeInvalidHeaderSys,
	"not found headers from ctx",
)

func getNotFoundKeyErr(msg string) error {
	return errorx.CreateError(
		http.StatusForbidden,
		errorx.ErrCodeInvalidHeaderSys,
		msg,
	)
}

func GetUserIdAndCorpId(ctx context.Context) (userId, corpId uint64) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		corpId, _ = h[CorpIdKey].(uint64)
		userId, _ = h[UserIdKey].(uint64)
	}
	return
}
func GetCorpId(ctx context.Context) (corpId uint64) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		corpId, _ = h[CorpIdKey].(uint64)
	}
	return
}
func GetUserId(ctx context.Context) (userId uint64) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		userId, _ = h[UserIdKey].(uint64)
	}
	return
}

func GetClientId(ctx context.Context) (clientId string) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		clientId, _ = h[ClientIdKey].(string)
	}
	return
}

func GetClientType(ctx context.Context) (clientType string) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		clientType, _ = h[ClientTypeKey].(string)
	}
	return
}
func GetRemoteIp(ctx context.Context) (ip string) {
	h, ok := ctx.Value(HeadersKey).(Header)
	if ok {
		ip, _ = h[RemoteIp].(string)
	}
	return
}

func MustCorpId(ctx context.Context) (corpId uint64, err error) {
	var ok bool
	h, ok := ctx.Value(HeadersKey).(Header)
	if !ok {
		return 0, NotFoundHeaders
	}

	corpId, ok = h[CorpIdKey].(uint64)
	if !ok {
		return 0, getNotFoundKeyErr("not found corp_id from headers")
	}
	return
}
func MustUserId(ctx context.Context) (userId uint64, err error) {
	var ok bool
	h, ok := ctx.Value(HeadersKey).(Header)
	if !ok {
		return 0, NotFoundHeaders
	}
	userId, ok = h[UserIdKey].(uint64)
	if !ok {
		return 0, getNotFoundKeyErr("not found user_id from headers")
	}
	return
}
func MustClientId(ctx context.Context) (clientId string, err error) {
	var ok bool
	h, ok := ctx.Value(HeadersKey).(Header)
	if !ok {
		return "", NotFoundHeaders
	}
	clientId, ok = h[ClientIdKey].(string)
	if !ok {
		return "", getNotFoundKeyErr("not found client_id from headers")
	}
	return
}
