package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// --------------------------------------------------------------------------
// Context keys
// --------------------------------------------------------------------------

type contextKey int

const (
	claimsCtxKey contextKey = iota
	userCtxKey
)

// --------------------------------------------------------------------------
// Claims
// --------------------------------------------------------------------------

// CognitoClaims represents JWT claims issued by AWS Cognito.
type CognitoClaims struct {
	Sub        string   `json:"sub"`
	CustomID   string   `json:"custom_id"`
	CustomRole string   `json:"custom_role"`
	TokenUse   string   `json:"token_use"` // "access" or "id"
	ClientID   string   `json:"client_id"` // present in access tokens
	Groups     []string `json:"cognito:groups"`
	jwt.RegisteredClaims
}

// --------------------------------------------------------------------------
// Config
// --------------------------------------------------------------------------

// CognitoConfig holds Cognito pool settings used to validate tokens.
type CognitoConfig struct {
	Region     string // e.g. "us-east-1"
	UserPoolID string // e.g. "us-east-1_xxxxxxxxx"
	ClientID   string // App client ID — leave empty to skip audience check
	// TokenUse restricts accepted token type: "access", "id", or "" (both).
	TokenUse string
}

func (c CognitoConfig) issuerURL() string {
	return fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", c.Region, c.UserPoolID)
}

func (c CognitoConfig) jwksURL() string {
	return c.issuerURL() + "/.well-known/jwks.json"
}

// --------------------------------------------------------------------------
// JWKS cache
// --------------------------------------------------------------------------

type jwksKey struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type keyCache struct {
	mu        sync.RWMutex
	keys      map[string]jwksKey
	fetchedAt time.Time
	ttl       time.Duration
}

func newKeyCache() *keyCache {
	return &keyCache{
		keys: make(map[string]jwksKey),
		ttl:  time.Hour,
	}
}

func (kc *keyCache) get(kid, jwksURL string) (*rsa.PublicKey, error) {
	// Optimistic read
	kc.mu.RLock()
	key, ok := kc.keys[kid]
	stale := time.Since(kc.fetchedAt) > kc.ttl
	kc.mu.RUnlock()

	if ok && !stale {
		return toRSAPublicKey(key)
	}

	kc.mu.Lock()
	defer kc.mu.Unlock()

	// Re-check under write lock before fetching (double-checked locking)
	if key, ok := kc.keys[kid]; ok && time.Since(kc.fetchedAt) <= kc.ttl {
		return toRSAPublicKey(key)
	}

	fetched, err := fetchJWKS(jwksURL)
	if err != nil {
		if ok { // serve stale
			return toRSAPublicKey(key)
		}
		return nil, fmt.Errorf("jwks fetch failed: %w", err)
	}

	kc.keys = fetched
	kc.fetchedAt = time.Now()

	newKey, found := kc.keys[kid]
	if !found {
		return nil, fmt.Errorf("no public key found for kid=%q", kid)
	}
	return toRSAPublicKey(newKey)
}

func fetchJWKS(url string) (map[string]jwksKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks endpoint returned %d", resp.StatusCode)
	}

	var payload struct {
		Keys []jwksKey `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	m := make(map[string]jwksKey, len(payload.Keys))
	for _, k := range payload.Keys {
		m[k.Kid] = k
	}
	return m, nil
}

// toRSAPublicKey converts a JWK into an *rsa.PublicKey without any PEM round-trip.
func toRSAPublicKey(k jwksKey) (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %q", k.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

// --------------------------------------------------------------------------
// Middleware constructors
// --------------------------------------------------------------------------

// CognitoAuth validates Cognito JWTs. Aborts with 401 on any failure.
func CognitoAuth(cfg CognitoConfig) gin.HandlerFunc {
	return buildMiddleware(cfg, newKeyCache(), true)
}

// OptionalCognitoAuth validates Cognito JWTs but does NOT abort when no token
// is present. Useful for endpoints serving both authenticated and anonymous users.
func OptionalCognitoAuth(cfg CognitoConfig) gin.HandlerFunc {
	return buildMiddleware(cfg, newKeyCache(), false)
}

func buildMiddleware(cfg CognitoConfig, kc *keyCache, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := extractBearer(c)
		if err != nil {
			if required {
				abortUnauthorized(c, "AUTH_TOKEN_INVALID", err.Error())
				return
			}
			c.Next()
			return
		}

		claims, err := parseAndValidate(tokenStr, cfg, kc)
		if err != nil {
			if required {
				abortUnauthorized(c, "AUTH_TOKEN_INVALID", err.Error())
				return
			}
			c.Next()
			return
		}

		c.Set(string(claimsCtxKey), claims)
		c.Set(string(userCtxKey), User{
			id:   claims.CustomID,
			role: claims.CustomRole,
		})
		c.Next()
	}
}

// RequireGroup middleware ensures the authenticated user belongs to at least
// one of the listed Cognito groups. Must be placed after CognitoAuth.
func RequireCognitoGroup(groups ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		allowed[g] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := ClaimsFromCtx(c)
		if !ok {
			abortUnauthorized(c, "AUTH_UNAUTHORIZED", "authentication required")
			return
		}
		for _, g := range claims.Groups {
			if _, ok := allowed[g]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
			"code":    "AUTH_FORBIDDEN",
			"message": "insufficient permissions",
		}})
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := ClaimsFromCtx(c)
		if !ok {
			abortUnauthorized(c, "AUTH_UNAUTHORIZED", "authentication required")
			return
		}
		if _, ok := allowed[claims.CustomRole]; ok {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
			"code":    "AUTH_FORBIDDEN",
			"message": "insufficient permissions",
		}})
	}
}

// --------------------------------------------------------------------------
// Token parsing & validation
// --------------------------------------------------------------------------

func parseAndValidate(tokenStr string, cfg CognitoConfig, kc *keyCache) (*CognitoClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&CognitoClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid in token header")
			}
			return kc.get(kid, cfg.jwksURL())
		},
		jwt.WithIssuer(cfg.issuerURL()),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*CognitoClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate token_use
	if cfg.TokenUse != "" && claims.TokenUse != cfg.TokenUse {
		return nil, fmt.Errorf("expected token_use=%q, got %q", cfg.TokenUse, claims.TokenUse)
	}

	// Validate client_id / audience
	if cfg.ClientID != "" {
		switch claims.TokenUse {
		case "access":
			if claims.ClientID != cfg.ClientID {
				return nil, errors.New("client_id mismatch")
			}
		case "id":
			aud, err := claims.GetAudience()
			if err != nil || !sliceContains(aud, cfg.ClientID) {
				return nil, errors.New("audience mismatch")
			}
		}
	}

	return claims, nil
}

// --------------------------------------------------------------------------
// Context helpers (call from handlers)
// --------------------------------------------------------------------------

// GetClaims retrieves CognitoClaims stored by the middleware.
func ClaimsFromCtx(c *gin.Context) (*CognitoClaims, bool) {
	v, ok := c.Get(string(claimsCtxKey))
	if !ok {
		return nil, false
	}
	claims, ok := v.(*CognitoClaims)
	return claims, ok
}

// GetUser returns the User from context.
func UserFromCtx(c *gin.Context) (User, bool) {
	v, ok := c.Get(string(userCtxKey))
	if !ok {
		return User{}, false
	}
	u, ok := v.(User)
	return u, ok
}

// --------------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------------

func extractBearer(c *gin.Context) (string, error) {
	h := c.GetHeader("Authorization")
	if h == "" {
		return "", errors.New("no Authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", errors.New("authorization header must be: Bearer <token>")
	}
	t := strings.TrimSpace(parts[1])
	if t == "" {
		return "", errors.New("empty bearer token")
	}
	return t, nil
}

func abortUnauthorized(c *gin.Context, code, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]string{
		"code":     code,
		"mesasage": msg,
	}})
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
