package middleware

import (
	"context"

	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID
	Role   string
}

type ctxKey string

const claimsKey ctxKey = "claims"

func SetClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func GetClaims(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}
