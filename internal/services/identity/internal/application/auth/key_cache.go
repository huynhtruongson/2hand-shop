package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

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
