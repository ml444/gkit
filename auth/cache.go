package auth

import (
	"context"
	"encoding/json"
	"git.csautodriver.com/base/cache"
	"git.csautodriver.com/base/gkit/auth/jwt"
	log "github.com/ml444/glog"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	JWTAuthDataHashKey = "jwt_auth_data"
	//JWTAuthBlackList   = "jwt_auth_black_list"

	expire = time.Hour * 24
)

var redisCli *redis.Client
var disableRedis bool

func init() {
	log.Info("======>>> init redis <<<======")
	var err error
	redisCli, err = cache.GetRedisCli(nil)
	if err != nil {
		log.Errorf("redis err: %v", err)
		disableRedis = true
		return
	}
	err = redisCli.Ping(context.Background()).Err()
	if err != nil {
		log.Errorf("redis err: %v", err)
		disableRedis = true
	}

}

func ExistCacheAuthDataBySign(ctx context.Context, sign string) (bool, error) {
	if disableRedis {
		return false, nil
	}
	return redisCli.HExists(ctx, JWTAuthDataHashKey, sign).Result()
}

func GetCacheAuthDataBySign(ctx context.Context, sign string) (*jwt.CustomData, error) {
	strCmd := redisCli.HGet(ctx, JWTAuthDataHashKey, sign)
	dataByte, err := strCmd.Bytes()
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	var data jwt.CustomData
	err = json.Unmarshal(dataByte, &data)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	return &data, nil
}

// SetCacheAuthDataBySign 不返回错误，不影响主流程
func SetCacheAuthDataBySign(ctx context.Context, sign string, data *jwt.CustomData) {
	if disableRedis {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("err: %v", err)
		//return err
	}
	dataStr := string(b)
	intCmd := redisCli.HSet(ctx, JWTAuthDataHashKey, sign, dataStr)
	if n, err := intCmd.Result(); err != nil {
		// 不返回错误
		log.Error(err.Error())
	} else if n == 0 {
		log.Warn("[HSET %s %s %s] failed", JWTAuthDataHashKey, sign, dataStr)
	}
	log.Debugf("HSET %s %s %s", JWTAuthDataHashKey, sign, dataStr)
}

func DelCacheAuthDataBySign(ctx context.Context, sign string) {
	err := redisCli.HDel(ctx, JWTAuthDataHashKey, sign).Err()
	if err != nil {
		log.Errorf("err: %v", err)
	}
}

func PutToAuthBlackList(ctx context.Context, clientId string, value string) error {
	return redisCli.Set(ctx, clientId, value, expire).Err()
}

func ExistFormAuthBlackList(ctx context.Context, clientId string) (bool, error) {
	if disableRedis {
		return false, nil
	}
	_, err := redisCli.Get(ctx, clientId).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
