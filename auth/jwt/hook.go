package jwt

import "context"

type HookFunc func(ctx context.Context, claims *CustomClaims) error
