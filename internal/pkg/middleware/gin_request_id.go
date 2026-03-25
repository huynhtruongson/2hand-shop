package middleware

import (
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

		// Make it available to all downstream handlers and middleware
		c.Set(RequestIDKey, requestID)

		// Echo it back so clients can correlate with their own logs
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	id, _ := c.Get(RequestIDKey)
	requestID, _ := id.(string)
	return requestID
}
