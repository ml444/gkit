package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ml444/gkit/errorx"
)

const (
	JWTClaimsKey = "JWTClaims"
)

const (
	bearer       string = "bearer"
	BearerPrefix string = "Bearer "
	bearerFormat string = "Bearer %s"
)

type RegisteredClaims = jwt.RegisteredClaims
type NumericDate = jwt.NumericDate
type CustomClaims struct {
	CustomData map[string]interface{} `json:"custom_data"`
	RegisteredClaims
}

func GenerateJWT(claims CustomClaims, secretKey []byte) (string, error) {
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(bearerFormat, tokenString), nil
}

func ParseJWT(tokenString string, secret []byte) (*CustomClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithLeeway(time.Second*5))
	if err != nil {
		return nil, errorx.CreateError(http.StatusUnauthorized, errorx.ErrCodeInvalidHeaderSys, err.Error())
	}

	// Get custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		err = errorx.CreateError(
			http.StatusUnauthorized,
			errorx.ErrCodeInvalidHeaderSys,
			"Claims assertion failure",
		)
		return nil, err
	}
	return claims, nil
}

func ParseJWT2ContextByHTTP(ctx context.Context, r *http.Request, secret []byte) error {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], bearer) {
		return errorx.CreateError(
			http.StatusPaymentRequired,
			errorx.ErrCodeInvalidHeaderSys,
			"The Authorization token is incorrectly formatted",
		)
	}
	tokenString := authHeaderParts[1]

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithLeeway(time.Second*5))
	if err != nil {
		return errorx.CreateError(http.StatusUnauthorized, errorx.ErrCodeInvalidHeaderSys, err.Error())
	}

	// Get custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		err = errorx.CreateError(
			http.StatusUnauthorized,
			errorx.ErrCodeInvalidHeaderSys,
			"Claims assertion failure",
		)
		return err
	}
	ctx = context.WithValue(ctx, JWTClaimsKey, claims)
	return nil
}

func AddJWT2HttpHeader(token string, r *http.Request) {
	if !strings.HasPrefix(token, BearerPrefix) {
		token = fmt.Sprintf(bearerFormat, token)
	}
	r.Header.Add("Authorization", token)
}
