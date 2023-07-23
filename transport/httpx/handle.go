package httpx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ml444/gutil/netx"
	"github.com/ml444/gutil/typex"
	"google.golang.org/grpc/metadata"

	"github.com/ml444/gkit/auth"
	"github.com/ml444/gkit/auth/jwt"
	"github.com/ml444/gkit/biz/header"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/pkg/env"
)

func getCacheTokenCustomData(ctx context.Context, signatureStr string) (*jwt.CustomData, error) {
	// cache
	exist, err := auth.ExistCacheAuthDataBySign(ctx, signatureStr)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	var data *jwt.CustomData
	if exist {
		data, err = auth.GetCacheAuthDataBySign(ctx, signatureStr)
		if err != nil {
			log.Errorf("err: %v", err)
			return nil, err
		}
	}
	return data, nil
}

func transferHeaderToCtx(ctx context.Context, r *http.Request, hook jwt.HookFunc, isTransferToken bool) (context.Context, error) {
	h := map[string]string{
		header.ClientTypeKey: header.GetHeader4HTTP(r.Header, header.ClientTypeKey),
		header.RemoteIpKey:   netx.GetRemoteIp(r),
		header.HttpPathKey:   r.URL.Path,
		header.TraceIdKey:    header.GetTraceId4Headers(r.Header),
	}
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Warn("not found Authorization")
		if env.IsLocalEnv() {
			h[header.CorpIdKey] = header.GetHeader4HTTP(r.Header, header.CorpIdKey)
			h[header.UserIdKey] = header.GetHeader4HTTP(r.Header, header.UserIdKey)
			h[header.ClientIdKey] = header.GetHeader4HTTP(r.Header, header.ClientIdKey)
		}
		md := metadata.New(h)
		ctx = metadata.NewIncomingContext(ctx, md)
		return ctx, nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], jwt.Bearer) {
		return ctx, jwt.ErrTokenFormat()
	}
	tokenString := authHeaderParts[1]
	if isTransferToken {
		h[header.TokenKey] = tokenString
	}
	partList := strings.Split(tokenString, ".")
	if len(partList) != 3 {
		return ctx, jwt.ErrTokenFormat()
	}
	signatureStr := partList[2]
	data, err := getCacheTokenCustomData(ctx, signatureStr)
	if err != nil {
		log.Errorf("err: %v", err)
		return ctx, err
	}
	var needSaveData bool
	if data == nil {
		var claims *jwt.CustomClaims
		claims, err = jwt.ParsePayload(partList[1])
		if err != nil {
			log.Error(err)
			return ctx, err
		}
		//ctx = context.WithValue(ctx, jwt.KeyJWTClaims, claims)
		if hook != nil {
			err = hook(ctx, claims)
			if err != nil {
				log.Error(err)
				return ctx, err
			}
		}

		// set cache
		data = &claims.CustomData
		data.ClientId = claims.ID
		needSaveData = true
	}
	if data != nil {
		// 验证token是否已经退出登录
		var exist bool
		exist, err = auth.ExistFormAuthBlackList(ctx, data.ClientId)
		if err != nil {
			log.Errorf("err: %v", err)
			return ctx, err
		}
		if exist {
			auth.DelCacheAuthDataBySign(ctx, signatureStr)
			return ctx, errorx.CreateError(http.StatusForbidden, errorx.ErrCodeInvalidHeaderSys, "token has expired")
		}
		if needSaveData {
			auth.SetCacheAuthDataBySign(ctx, signatureStr, data)
		}
		log.Debug(data)
		h[header.CorpIdKey] = strconv.FormatUint(data.CorpId, 10)
		h[header.UserIdKey] = strconv.FormatUint(data.UserId, 10)
		h[header.ClientIdKey] = data.ClientId
		h[header.ClientTypeKey] = data.ClientType
		if data.Extra != nil {
			for k, v := range data.Extra {
				h[k] = typex.AnyToStr(v)
			}
		}
	}

	md := metadata.New(h)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx, nil
}

func AddJWT2HttpHeader(token string, r *http.Request) {
	if !strings.HasPrefix(token, jwt.BearerPrefix) {
		token = fmt.Sprintf(jwt.BearerFormat, token)
	}
	r.Header.Add("Authorization", token)
}
