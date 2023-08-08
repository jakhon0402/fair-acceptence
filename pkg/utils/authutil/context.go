package authutil

import (
	"context"
	"fajr-acceptance/internal/models"
	"github.com/gin-gonic/gin"
)

const (
	IdentityKey = "user.email"
)

type userContextKey string

const userKey = userContextKey("user")

func CurrentUser(ctx context.Context) *models.User {
	if ctx == nil {
		return nil
	}
	if p, ok := ctx.Value(userKey).(*models.User); ok {
		return p
	}
	if gctx, ok := ctx.(*gin.Context); ok && gctx != nil {
		return currentUserFromGinContext(gctx)
	}
	return nil
}

func currentUserFromGinContext(ctx *gin.Context) *models.User {
	v, ok := ctx.Get(IdentityKey)
	if !ok {
		return nil
	}
	if p, ok := v.(models.User); ok {
		return &p
	}
	return nil
}

// WithUserContext creates a new context with the provided user attached.
func WithUserContext(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}
