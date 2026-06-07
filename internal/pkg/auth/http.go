package auth

// import (
// 	"context"
// 	"crypto/rsa"
// 	"encoding/base64"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"math/big"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// )

// // --------------------------------------------------------------------------
// // Context keys
// // --------------------------------------------------------------------------

// type contextKey int

// const (
// 	claimsCtxKey contextKey = iota
// 	userCtxKey
// )

// // --------------------------------------------------------------------------
// // Claims
// // --------------------------------------------------------------------------

// // KeycloakClaims represents JWT claims issued by Keycloak.
// type KeycloakClaims struct {
// 	jwt.RegisteredClaims
// 	PreferredUsername string              `json:"preferred_username"`
// 	Email             string              `json:"email"`
// 	EmailVerified     bool                `json:"email_verified"`
// 	Name              string              `json:"name"`
// 	RealmAccess       realmAccess         `json:"realm_access"`
// 	ResourceAccess    map[string]roleList `json:"resource_access"`
// 	Scope             string              `json:"scope"`
// }

// // realmAccess holds realm-level roles embedded in the token.
// type realmAccess struct {
// 	Roles []string `json:"roles"`
// }

// // roleList holds per-resource roles embedded in resource_access.
// type roleList struct {
// 	Roles []string `json:"roles"`
// }

// // HasRealmRole reports whether the claims contain the given realm role.
// func (kc *KeycloakClaims) HasRealmRole(role string) bool {
// 	return sliceContains(kc.RealmAccess.Roles, role)
// }

// // HasClientRole reports whether the claims contain the given role for clientID.
// func (kc *KeycloakClaims) HasClientRole(clientID, role string) bool {
// 	res, ok := kc.ResourceAccess[clientID]
// 	if !ok {
// 		return false
// 	}
// 	return sliceContains(res.Roles, role)
// }

// // --------------------------------------------------------------------------
// // Config
// // --------------------------------------------------------------------------

// // KeycloakConfig holds the settings used to validate Keycloak-issued tokens.
// type KeycloakConfig struct {
// 	Env      string
// 	BaseURL  string // e.g. "https://auth.example.com"
// 	Realm    string // e.g. "my-realm"
// 	ClientID string // App client ID — used for audience / client-role checks
// }

// func (c KeycloakConfig) issuerURL() string {
// 	return fmt.Sprintf("%s/realms/%s", c.BaseURL, c.Realm)
// }

// func (c KeycloakConfig) jwksURL() string {
// 	return c.issuerURL() + "/protocol/openid-connect/certs"
// }

// // --------------------------------------------------------------------------
// // JWKS cache
// // --------------------------------------------------------------------------

// type jwksKey struct {
// 	Kid string `json:"kid"`
// 	Alg string `json:"alg"`
// 	Kty string `json:"kty"`
// 	Use string `json:"use"`
// 	N   string `json:"n"`
// 	E   string `json:"e"`
// }

// type keyCache struct {
// 	mu        sync.RWMutex
// 	keys      map[string]jwksKey
// 	fetchedAt time.Time
// 	ttl       time.Duration
// }

// func newKeyCache() *keyCache {
// 	return &keyCache{
// 		keys: make(map[string]jwksKey),
// 		ttl:  time.Hour,
// 	}
// }

// func (kc *keyCache) get(kid, jwksURL string) (*rsa.PublicKey, error) {
// 	// Optimistic read
// 	kc.mu.RLock()
// 	key, ok := kc.keys[kid]
// 	stale := time.Since(kc.fetchedAt) > kc.ttl
// 	kc.mu.RUnlock()

// 	if ok && !stale {
// 		return toRSAPublicKey(key)
// 	}

// 	kc.mu.Lock()
// 	defer kc.mu.Unlock()

// 	// Re-check under write lock before fetching (double-checked locking)
// 	if key, ok := kc.keys[kid]; ok && time.Since(kc.fetchedAt) <= kc.ttl {
// 		return toRSAPublicKey(key)
// 	}

// 	fetched, err := fetchJWKS(jwksURL)
// 	if err != nil {
// 		if ok { // serve stale
// 			return toRSAPublicKey(key)
// 		}
// 		return nil, fmt.Errorf("jwks fetch failed: %w", err)
// 	}

// 	kc.keys = fetched
// 	kc.fetchedAt = time.Now()

// 	newKey, found := kc.keys[kid]
// 	if !found {
// 		return nil, fmt.Errorf("no public key found for kid=%q", kid)
// 	}
// 	return toRSAPublicKey(newKey)
// }

// func fetchJWKS(url string) (map[string]jwksKey, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("jwks endpoint returned %d", resp.StatusCode)
// 	}

// 	var payload struct {
// 		Keys []jwksKey `json:"keys"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
// 		return nil, err
// 	}

// 	m := make(map[string]jwksKey, len(payload.Keys))
// 	for _, k := range payload.Keys {
// 		m[k.Kid] = k
// 	}
// 	return m, nil
// }

// // toRSAPublicKey converts a JWK into an *rsa.PublicKey without any PEM round-trip.
// func toRSAPublicKey(k jwksKey) (*rsa.PublicKey, error) {
// 	if k.Kty != "RSA" {
// 		return nil, fmt.Errorf("unsupported key type: %q", k.Kty)
// 	}

// 	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid modulus: %w", err)
// 	}
// 	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid exponent: %w", err)
// 	}

// 	n := new(big.Int).SetBytes(nBytes)
// 	e := int(new(big.Int).SetBytes(eBytes).Int64())

// 	return &rsa.PublicKey{N: n, E: e}, nil
// }

// // --------------------------------------------------------------------------
// // Middleware constructors
// // --------------------------------------------------------------------------

// // KeycloakAuth validates Keycloak JWTs. Aborts with 401 on any failure.
// func KeycloakAuth(cfg KeycloakConfig) gin.HandlerFunc {
// 	return buildMiddleware(cfg, newKeyCache(), true)
// }

// // OptionalKeycloakAuth validates Keycloak JWTs but does NOT abort when no token
// // is present. Useful for endpoints serving both authenticated and anonymous users.
// func OptionalKeycloakAuth(cfg KeycloakConfig) gin.HandlerFunc {
// 	return buildMiddleware(cfg, newKeyCache(), false)
// }

// func buildMiddleware(cfg KeycloakConfig, kc *keyCache, required bool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		tokenStr, err := extractBearer(c)
// 		if err != nil {
// 			if required {
// 				abortUnauthorized(c, "AUTH_TOKEN_MISSING", err.Error())
// 				return
// 			}
// 			c.Next()
// 			return
// 		}

// 		claims, err := parseAndValidate(tokenStr, cfg, kc)
// 		if err != nil {
// 			if required {
// 				abortUnauthorized(c, "AUTH_TOKEN_INVALID", err.Error())
// 				return
// 			}
// 			c.Next()
// 			return
// 		}

// 		c.Set(claimsCtxKey, claims)
// 		c.Set(userCtxKey, NewUserCtx(claims.Subject, claims.Email, claims.RealmAccess.Roles))
// 		c.Next()
// 	}
// }

// // RequireRealmRole ensures the authenticated user has at least one of the given
// // Keycloak realm roles. Must be placed after KeycloakAuth.
// func RequireRealmRole(roles ...string) gin.HandlerFunc {
// 	allowed := make(map[string]struct{}, len(roles))
// 	for _, r := range roles {
// 		allowed[r] = struct{}{}
// 	}

// 	return func(c *gin.Context) {
// 		claims, ok := ClaimsFromCtx(c)
// 		if !ok {
// 			abortUnauthorized(c, "AUTH_UNAUTHORIZED", "authentication required")
// 			return
// 		}
// 		for _, r := range claims.RealmAccess.Roles {
// 			if _, ok := allowed[r]; ok {
// 				c.Next()
// 				return
// 			}
// 		}
// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
// 			"code":    "AUTH_FORBIDDEN",
// 			"message": "insufficient realm roles",
// 		}})
// 	}
// }

// // RequireClientRole ensures the authenticated user has at least one of the given
// // roles for the specified Keycloak client. Must be placed after KeycloakAuth.
// func RequireClientRole(clientID string, roles ...string) gin.HandlerFunc {
// 	allowed := make(map[string]struct{}, len(roles))
// 	for _, r := range roles {
// 		allowed[r] = struct{}{}
// 	}

// 	return func(c *gin.Context) {
// 		claims, ok := ClaimsFromCtx(c)
// 		if !ok {
// 			abortUnauthorized(c, "AUTH_UNAUTHORIZED", "authentication required")
// 			return
// 		}
// 		res, found := claims.ResourceAccess[clientID]
// 		if found {
// 			for _, r := range res.Roles {
// 				if _, ok := allowed[r]; ok {
// 					c.Next()
// 					return
// 				}
// 			}
// 		}
// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": map[string]string{
// 			"code":    "AUTH_FORBIDDEN",
// 			"message": "insufficient client roles",
// 		}})
// 	}
// }

// // --------------------------------------------------------------------------
// // Token parsing & validation
// // --------------------------------------------------------------------------

// func parseAndValidate(tokenStr string, cfg KeycloakConfig, kc *keyCache) (*KeycloakClaims, error) {
// 	token, err := jwt.ParseWithClaims(
// 		tokenStr,
// 		&KeycloakClaims{},
// 		func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			kid, _ := token.Header["kid"].(string)
// 			if kid == "" {
// 				return nil, errors.New("missing kid in token header")
// 			}
// 			return kc.get(kid, cfg.jwksURL())
// 		},
// 		// jwt.WithIssuer(cfg.issuerURL()),
// 		jwt.WithExpirationRequired(),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("token validation failed: %w", err)
// 	}

// 	claims, ok := token.Claims.(*KeycloakClaims)
// 	if !ok || !token.Valid {
// 		return nil, errors.New("invalid token claims")
// 	}

// 	return claims, nil
// }

// // --------------------------------------------------------------------------
// // Context helpers (call from handlers)
// // --------------------------------------------------------------------------

// // ClaimsFromCtx retrieves KeycloakClaims stored by the middleware.
// func ClaimsFromCtx(c *gin.Context) (*KeycloakClaims, bool) {
// 	v, ok := c.Get(claimsCtxKey)
// 	if !ok {
// 		return nil, false
// 	}
// 	claims, ok := v.(*KeycloakClaims)
// 	return claims, ok
// }

// // GetUser returns the User from context.
// func UserFromCtx(c *gin.Context) (UserCtx, bool) {
// 	v, ok := c.Get(userCtxKey)
// 	if !ok {
// 		return UserCtx{}, false
// 	}
// 	u, ok := v.(UserCtx)
// 	return u, ok
// }

// // --------------------------------------------------------------------------
// // Internal helpers
// // --------------------------------------------------------------------------

// func extractBearer(c *gin.Context) (string, error) {
// 	h := c.GetHeader("Authorization")
// 	if h == "" {
// 		return "", errors.New("no Authorization header")
// 	}
// 	parts := strings.SplitN(h, " ", 2)
// 	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
// 		return "", errors.New("authorization header must be: Bearer <token>")
// 	}
// 	t := strings.TrimSpace(parts[1])
// 	if t == "" {
// 		return "", errors.New("empty bearer token")
// 	}
// 	return t, nil
// }

// func abortUnauthorized(c *gin.Context, code, msg string) {
// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]string{
// 		"code":    code,
// 		"message": msg,
// 	}})
// }

// func sliceContains(slice []string, s string) bool {
// 	for _, v := range slice {
// 		if v == s {
// 			return true
// 		}
// 	}
// 	return false
// }
