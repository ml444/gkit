package header

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ml444/gutil/typex"
	"google.golang.org/grpc/metadata"

	"github.com/ml444/gkit/errorx"
)

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
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[CorpIdKey]) > 0 {
			corpId, _ = strconv.ParseUint(md[CorpIdKey][0], 10, 64)
		}
		if len(md[UserIdKey]) > 0 {
			userId, _ = strconv.ParseUint(md[UserIdKey][0], 10, 64)
		}
	}
	return
}
func GetCorpId(ctx context.Context) (corpId uint64) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[CorpIdKey]) > 0 {
			corpId, _ = strconv.ParseUint(md[CorpIdKey][0], 10, 64)
		}
	}
	return
}
func GetUserId(ctx context.Context) (userId uint64) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[UserIdKey]) > 0 {
			userId, _ = strconv.ParseUint(md[UserIdKey][0], 10, 64)
		}
	}
	return
}

func GetClientId(ctx context.Context) (clientId string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[ClientIdKey]) > 0 {
			clientId = md[ClientIdKey][0]
		}
	}
	return
}

func GetClientType(ctx context.Context) (clientType string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[ClientTypeKey]) > 0 {
			clientType = md[ClientTypeKey][0]
		}
	}
	return
}
func GetRemoteIp(ctx context.Context) (ip string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[RemoteIpKey]) > 0 {
			ip = md[RemoteIpKey][0]
		}
	}
	return
}

func MustCorpId(ctx context.Context) (corpId uint64, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[CorpIdKey]) > 0 {
			return strconv.ParseUint(md[CorpIdKey][0], 10, 64)
		}
	} else {
		return 0, NotFoundHeaders
	}
	return 0, getNotFoundKeyErr("not found corp_id from headers")
}

func MustUserId(ctx context.Context) (userId uint64, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[UserIdKey]) > 0 {
			return strconv.ParseUint(md[UserIdKey][0], 10, 64)
		}
	} else {
		return 0, NotFoundHeaders
	}

	return 0, getNotFoundKeyErr("not found user_id from headers")
}

func MustClientId(ctx context.Context) (clientId string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[ClientIdKey]) > 0 {
			return md[ClientIdKey][0], nil
		}
	} else {
		return "", NotFoundHeaders
	}

	return "", getNotFoundKeyErr("not found client_id from headers")
}

func TransportMDFromCtx(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func GetTaskId(ctx context.Context) (taskId uint64) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[TaskIdKey]) > 0 {
			taskId, _ = strconv.ParseUint(md[TaskIdKey][0], 10, 64)
		}
	}
	return
}

func GetTraceId(ctx context.Context) (traceId string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if len(md[TraceIdKey]) > 0 {
			traceId = md[TraceIdKey][0]
		}
	}
	return
}

func GetHttpPath(ctx context.Context) (path string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md[HttpPathKey]) > 0 {
		path = md[HttpPathKey][0]
	}
	return
}

func SetValue(ctx context.Context, key, val string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md.Set(key, val)
	} else {
		md = metadata.MD{
			key: []string{val},
		}
	}
	return metadata.NewIncomingContext(ctx, md)
}

func NewCtx(optMap map[string]interface{}) context.Context {
	ctx := context.Background()
	md := metadata.MD{}
	for k, v := range optMap {
		md.Set(k, typex.AnyToStr(v))
	}
	return metadata.NewIncomingContext(ctx, md)
}

func NewCtxWithCorpAndUser(corpId, userId uint64) context.Context {
	ctx := context.Background()
	md := metadata.MD{
		CorpIdKey: []string{strconv.FormatUint(corpId, 10)},
		UserIdKey: []string{strconv.FormatUint(userId, 10)},
	}
	return metadata.NewIncomingContext(ctx, md)
}
