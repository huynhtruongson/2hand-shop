package stripe

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v85/webhook"
)

// WebhookMiddleware returns a Gin middleware that verifies Stripe webhook signatures
// and stores the parsed event in the Gin context under the key "stripe_event".
// It consumes and restores the request body so downstream handlers can still read it.
func WebhookMiddleware(webhookSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		// Restore body so downstream handlers can re-read it.
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		sigHeader := c.GetHeader("Stripe-Signature")
		event, err := webhook.ConstructEvent(body, sigHeader, webhookSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid stripe signature"})
			return
		}

		c.Set("stripe_event", event)
		c.Next()
	}
}
