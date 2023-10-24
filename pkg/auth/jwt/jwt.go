package jwt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
)

const (
	KeyJWTClaims = "JWTClaims"
	KeyJWTToken  = "JWTToken"
)

const (
	Bearer       string = "bearer"
	BearerPrefix string = "Bearer "
	BearerFormat string = "Bearer %s"
)

type RegisteredClaims = jwt.RegisteredClaims
type NumericDate = jwt.NumericDate
type CustomClaims struct {
	CustomData
	RegisteredClaims
}

type CustomData struct {
	CorpId     uint64                 `json:"corp_id,omitempty"`
	UserId     uint64                 `json:"user_id,omitempty"`
	ClientType string                 `json:"client_type,omitempty"`
	ClientId   string                 `json:"client_id,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

func GenerateJWT(claims CustomClaims, secretKey []byte) (string, error) {
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(secretKey)
	if err != nil {
		log.Errorf("err: %v", err)
		return "", err
	}
	// Sign token with secret key
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return fmt.Sprintf(BearerFormat, tokenString), nil
}

func ParseToken(tokenString string, secret []byte) (*CustomClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithLeeway(time.Second*5),
		//jwt.WithoutClaimsValidation(),
	)
	if err != nil {
		log.Error(err)
		return nil, errorx.CreateError(http.StatusUnauthorized, errorx.ErrCodeInvalidHeaderSys, err.Error())
	}

	// Get custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrClaims()
	}
	return claims, nil
}

func ParsePayload(payload string) (*CustomClaims, error) {
	payloadByte, err := jwt.NewParser().DecodeSegment(payload)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	var claims CustomClaims
	err = json.Unmarshal(payloadByte, &claims)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	return &claims, nil
}
