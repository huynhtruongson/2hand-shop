package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

type requestIDKey int

const (
	RequestIDHeader              = "X-Request-ID"
	RequestIDKey    requestIDKey = iota
)

func GinRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}

		ctx := context.WithValue(c.Request.Context(), "reqid", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

func getRequestID(c *gin.Context) string {
	id, _ := c.Get(RequestIDKey)
	requestID, _ := id.(string)
	return requestID
}

func GetRequestIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value("reqid").(string); ok {
		return id
	}
	return ""
}
