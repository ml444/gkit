package auth

import (
	"fmt"
	"github.com/ml444/gkit/auth/jwt"
	"github.com/ml444/gkit/log"
	"os"
	"time"
)

const (
	IssuerName = "cs110@csautodriver.com"

	EnvKeySecretFilePath = "AUTH_JWT_PEM"
)

func getKid(userId uint64) string {
	return fmt.Sprintf("%d_%d", userId, time.Now().UnixMilli())
}
func GenerateJWT(userId, corpId uint64, clientType string) (string, error) {
	secret, err := getAuthSecret()
	if err != nil {
		log.Errorf("err: %v", err)
		return "", err
	}
	// Set custom claims
	claims := jwt.CustomClaims{
		CustomData: jwt.CustomData{
			CorpId:     corpId,
			UserId:     userId,
			ClientType: clientType,
		},

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(expire)}, // Expires in 24 hours
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},             // 签发时间
			NotBefore: &jwt.NumericDate{Time: time.Now()},             // 生效时间
			Issuer:    IssuerName,
			//Subject:   "somebody",
			ID: getKid(userId),
			//Audience:  []string{"somebody_else"},
		},
	}

	tokenString, err := jwt.GenerateJWT(claims, secret)
	if err != nil {
		log.Errorf("err: %v", err)
		return "", err
	}

	return tokenString, nil
}

func getAuthSecret() ([]byte, error) {
	return os.ReadFile(os.Getenv(EnvKeySecretFilePath))
}

func ParseToken(token string) (*jwt.CustomClaims, error) {
	secret, err := getAuthSecret()
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	return jwt.ParseToken(token, secret)
}
