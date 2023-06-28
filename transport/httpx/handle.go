package httpx

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ml444/gkit/auth"
	"github.com/ml444/gkit/auth/jwt"
	"github.com/ml444/gkit/core"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gutil/netx"
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

func HandleContextByHTTP(ctx context.Context, r *http.Request, hook jwt.HookFunc) (context.Context, error) {
	header := core.Header{
		core.ClientTypeKey: core.GetHeader4HTTP(r.Header, core.HttpHeaderClientType),
		core.RemoteIp:      netx.GetRemoteIp(r),
	}
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Warn("not found Authorization")
		if core.IsLocalEnv() {
			header[core.CorpIdKey] = core.GetCorpId4Headers(r.Header)
			header[core.UserIdKey] = core.GetUserId4Headers(r.Header)
			header[core.ClientIdKey] = core.GetHeader4HTTP(r.Header, core.HttpHeaderClientId)
		}
		ctx = context.WithValue(ctx, core.HeadersKey, header)
		return ctx, nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], jwt.Bearer) {
		return ctx, jwt.ErrTokenFormat()
	}
	tokenString := authHeaderParts[1]

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
		log.Info(data)
		header[core.CorpIdKey] = data.CorpId
		header[core.UserIdKey] = data.UserId
		header[core.ClientIdKey] = data.ClientId
		header[core.ClientTypeKey] = data.ClientType
		if data.Extra != nil {
			for k, v := range data.Extra {
				header[k] = v
			}
		}
	}
	ctx = context.WithValue(ctx, core.HeadersKey, header)

	//ctx = context.WithValue(ctx, jwt.KeyJWTToken, tokenString)
	return ctx, nil
}

func AddJWT2HttpHeader(token string, r *http.Request) {
	if !strings.HasPrefix(token, jwt.BearerPrefix) {
		token = fmt.Sprintf(jwt.BearerFormat, token)
	}
	r.Header.Add("Authorization", token)
}
