package middleware

import (
	"context"
	"fajr-acceptance/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

const (
	XRequestIdKey = "X-Request-ID"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get(XRequestIdKey)
		if requestId == "" {
			requestId = uuid.New().String()
		}

		logging := logger.NewLogger().WithField("requestId", requestId).Logger
		c.Request = c.Request.WithContext(logger.WithLogger(c.Request.Context(), logging))
		c.Writer.Header().Set(XRequestIdKey, requestId)
		c.Next()
	}
}

// TimeoutMiddleware attach deadline to gin.Request.Context
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatus(http.StatusGatewayTimeout)
			}
			cancel()
		}()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
