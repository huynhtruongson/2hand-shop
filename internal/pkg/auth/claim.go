package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HeaderEmail  = "X-Auth-Request-Email"
	HeaderRoles  = "X-Auth-Request-Roles"
	HeaderUserID = "X-Auth-Request-User-Id"
)

// Claims holds all identity headers forwarded from the auth-service.
type Claims struct {
	userID string // sub claim — Keycloak internal user ID
	email  string // email
	roles  string // comma-separated realm roles e.g. "admin,client"
}

// GetClaims extracts all auth headers from the Gin context into a Claims struct.
// Call this once per handler and pass Claims around — avoids repeated
// c.GetHeader calls scattered across your codebase.
func GetClaims(c *gin.Context) Claims {
	return Claims{
		userID: c.GetHeader(HeaderUserID),
		email:  c.GetHeader(HeaderEmail),
		roles:  c.GetHeader(HeaderRoles),
	}
}

func (c Claims) UserID() string {
	return c.userID
}
func (c Claims) Email() string {
	return c.email
}
func (c Claims) Roles() string {
	return c.roles
}

func (c Claims) HasRole(role string) bool {
	if c.roles == "" {
		return false
	}
	for _, r := range splitRoles(c.roles) {
		if r == role {
			return true
		}
	}
	return false
}

func (c Claims) HasAdminRole() bool {
	return c.HasRole("admin")
}

func (c Claims) HasClientRole() bool {
	return c.HasRole("client")
}

func RequireRoleMiddleware(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roles := c.GetHeader(HeaderRoles)
		if roles == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
				"code":    "AUTH_FORBIDDEN",
				"message": "insufficient roles",
			}})
			return
		}
		for _, r := range splitRoles(roles) {
			if _, ok := allowed[r]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
			"code":    "AUTH_FORBIDDEN",
			"message": "insufficient roles",
		}})
	}
}

func splitRoles(roles string) []string {
	var out []string
	start := 0
	for i := 0; i < len(roles); i++ {
		if roles[i] == ',' {
			if s := roles[start:i]; s != "" {
				out = append(out, s)
			}
			start = i + 1
		}
	}
	if s := roles[start:]; s != "" {
		out = append(out, s)
	}
	return out
}
