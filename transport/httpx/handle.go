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

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/pkg/auth"
	jwt2 "github.com/ml444/gkit/pkg/auth/jwt"
	"github.com/ml444/gkit/pkg/env"
	header2 "github.com/ml444/gkit/pkg/header"
)

func getCacheTokenCustomData(ctx context.Context, signatureStr string) (*jwt2.CustomData, error) {
	// cache
	exist, err := auth.ExistCacheAuthDataBySign(ctx, signatureStr)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	var data *jwt2.CustomData
	if exist {
		data, err = auth.GetCacheAuthDataBySign(ctx, signatureStr)
		if err != nil {
			log.Errorf("err: %v", err)
			return nil, err
		}
	}
	return data, nil
}

func transferHeaderToCtx(ctx context.Context, r *http.Request, hook jwt2.HookFunc, isTransferToken bool) (context.Context, error) {
	h := map[string]string{
		header2.ClientTypeKey: header2.GetHeader4HTTP(r.Header, header2.ClientTypeKey),
		header2.RemoteIpKey:   netx.GetRemoteIp(r),
		header2.HttpPathKey:   r.URL.Path,
		header2.TraceIdKey:    header2.GetTraceId4Headers(r.Header),
	}
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Warn("not found Authorization")
		if env.IsLocalEnv() {
			h[header2.CorpIdKey] = header2.GetHeader4HTTP(r.Header, header2.CorpIdKey)
			h[header2.UserIdKey] = header2.GetHeader4HTTP(r.Header, header2.UserIdKey)
			h[header2.ClientIdKey] = header2.GetHeader4HTTP(r.Header, header2.ClientIdKey)
		}
		md := metadata.New(h)
		ctx = metadata.NewIncomingContext(ctx, md)
		return ctx, nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], jwt2.Bearer) {
		return ctx, jwt2.ErrTokenFormat()
	}
	tokenString := authHeaderParts[1]
	if isTransferToken {
		h[header2.TokenKey] = tokenString
	}
	partList := strings.Split(tokenString, ".")
	if len(partList) != 3 {
		return ctx, jwt2.ErrTokenFormat()
	}
	signatureStr := partList[2]
	data, err := getCacheTokenCustomData(ctx, signatureStr)
	if err != nil {
		log.Errorf("err: %v", err)
		return ctx, err
	}
	var needSaveData bool
	if data == nil {
		var claims *jwt2.CustomClaims
		claims, err = jwt2.ParsePayload(partList[1])
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
		h[header2.CorpIdKey] = strconv.FormatUint(data.CorpId, 10)
		h[header2.UserIdKey] = strconv.FormatUint(data.UserId, 10)
		h[header2.ClientIdKey] = data.ClientId
		h[header2.ClientTypeKey] = data.ClientType
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
	if !strings.HasPrefix(token, jwt2.BearerPrefix) {
		token = fmt.Sprintf(jwt2.BearerFormat, token)
	}
	r.Header.Add("Authorization", token)
}
