package migrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
)

// ReindexHelper provides low-level primitives for zero-downtime index management
// using ES aliases. All methods are additive — none of them delete indices
// automatically. Responsibility for cleanup belongs to the Down migration, which
// should call DeleteIndex after confirming the old concrete index is out of the
// read path via SwapAlias.
//
// For large indices, prefer calling ReindexInPlace from a pre-deploy CLI job rather
// than at application startup, as WaitForCompletion=true blocks until reindex
// completes.
type ReindexHelper struct {
	es *elasticsearch.Client
	lg logger.Logger
}

// NewReindexHelper constructs a ReindexHelper with the given ES client and logger.
func NewReindexHelper(es *elasticsearch.Client, lg logger.Logger) *ReindexHelper {
	return &ReindexHelper{es: es, lg: lg}
}

// CreateIndexWithAlias creates a concrete index with the provided mapping body,
// then atomically attaches an alias to it via the _aliases API. This is the
// recommended way to create a new index version that should be immediately
// readable via a stable alias.
//
// Use this in an Up migration to stand up a new concrete index without
// disrupting reads through the alias.
func (h *ReindexHelper) CreateIndexWithAlias(ctx context.Context, indexName, aliasName string, body io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 1. Create the concrete index.
	createRes, err := h.es.Indices.Create(
		indexName,
		h.es.Indices.Create.WithBody(body),
		h.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create index %s: %w", indexName, err)
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		respBody, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("create index %s: %s — %s", indexName, createRes.Status(), string(respBody))
	}

	// 2. Atomically attach the alias.
	if err := h.addAlias(ctx, indexName, aliasName); err != nil {
		return fmt.Errorf("attach alias %s to %s: %w", aliasName, indexName, err)
	}

	h.lg.Info("index created with alias", "index", indexName, "alias", aliasName)
	return nil
}

// ReindexInPlace copies all documents from sourceIndex to destIndex synchronously.
// It uses ES's Reindex API with WaitForCompletion=true, meaning this call blocks
// until the reindex operation fully completes.
//
// For large indices (millions of documents), consider running this as a background
// pre-deploy job instead, using WaitForCompletion=false to get a task ID that can
// be polled separately.
func (h *ReindexHelper) ReindexInPlace(ctx context.Context, sourceIndex, destIndex string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	body := map[string]any{
		"source": map[string]any{"index": sourceIndex},
		"dest":   map[string]any{"index": destIndex},
	}
	bodyBytes, _ := json.Marshal(body)

	res, err := h.es.Reindex(
		bytes.NewReader(bodyBytes),
		h.es.Reindex.WithContext(ctx),
		h.es.Reindex.WithWaitForCompletion(true),
	)
	if err != nil {
		return fmt.Errorf("reindex %s → %s: %w", sourceIndex, destIndex, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("reindex %s → %s: %s — %s", sourceIndex, destIndex, res.Status(), string(respBody))
	}

	var result struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode reindex response: %w", err)
	}

	if result.Failed > 0 {
		h.lg.Warn("reindex completed with failures",
			"total", result.Total, "successful", result.Successful, "failed", result.Failed)
		return fmt.Errorf("reindex %s → %s: %d documents failed", sourceIndex, destIndex, result.Failed)
	}

	h.lg.Info("reindex completed", "total", result.Total, "successful", result.Successful)
	return nil
}

// SwapAlias atomically moves aliasName from oldIndex to newIndex using the
// _aliases API. Both indices must exist. This is the zero-downtime switch:
// reads and writes through aliasName instantly route to newIndex with no
// service interruption.
//
// Use this in an Up migration after ReindexInPlace to promote the new concrete
// index into the read path.
func (h *ReindexHelper) SwapAlias(ctx context.Context, aliasName, oldIndex, newIndex string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Strip any leading "/" that may have been included by mistake.
	aliasName = strings.TrimPrefix(aliasName, "/")
	oldIndex = strings.TrimPrefix(oldIndex, "/")
	newIndex = strings.TrimPrefix(newIndex, "/")

	action := map[string]any{
		"actions": []map[string]any{
			{"remove": map[string]any{"index": oldIndex, "alias": aliasName}},
			{"add": map[string]any{"index": newIndex, "alias": aliasName}},
		},
	}
	body, _ := json.Marshal(action)

	res, err := h.es.Indices.UpdateAliases(
		bytes.NewReader(body),
		h.es.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("swap alias %s: %s → %s: %w", aliasName, oldIndex, newIndex, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("swap alias %s: %s → %s: %s — %s", aliasName, oldIndex, newIndex, res.Status(), string(respBody))
	}

	h.lg.Info("alias swapped", "alias", aliasName, "from", oldIndex, "to", newIndex)
	return nil
}

// DeleteIndex deletes the named index. It returns nil if the index does not
// exist (404). Call this in a Down migration only after SwapAlias has
// confirmed the index is no longer referenced by any alias in the read path.
func (h *ReindexHelper) DeleteIndex(ctx context.Context, indexName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexName = strings.TrimPrefix(indexName, "/")

	res, err := h.es.Indices.Delete(
		[]string{indexName},
		h.es.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("delete index %s: %w", indexName, err)
	}
	defer res.Body.Close()
	// 404 means the index is already gone — treat as success.
	if res.IsError() && res.StatusCode != 404 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete index %s: %s — %s", indexName, res.Status(), string(respBody))
	}

	h.lg.Info("index deleted", "index", indexName)
	return nil
}

// IndexExists returns true if an index or alias with the given name exists.
// The ES Indices.Exists API returns 200 for both concrete indices and aliases.
func (h *ReindexHelper) IndexExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	name = strings.TrimPrefix(name, "/")

	res, err := h.es.Indices.Exists([]string{name}, h.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("check index exists %s: %w", name, err)
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}

// CurrentAliasTarget resolves which concrete index aliasName currently points to.
// If aliasName does not exist, it returns ("", nil).
func (h *ReindexHelper) CurrentAliasTarget(ctx context.Context, aliasName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	aliasName = strings.TrimPrefix(aliasName, "/")

	res, err := h.es.Indices.GetAlias(
		h.es.Indices.GetAlias.WithContext(ctx),
		h.es.Indices.GetAlias.WithName(aliasName),
	)
	if err != nil {
		return "", fmt.Errorf("get alias %s: %w", aliasName, err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return "", nil
	}
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("get alias %s: %s — %s", aliasName, res.Status(), string(respBody))
	}

	var indices map[string]struct {
		Aliases map[string]struct{} `json:"aliases"`
	}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return "", fmt.Errorf("decode alias response for %s: %w", aliasName, err)
	}

	for name := range indices {
		return name, nil
	}
	return "", nil
}

// addAlias adds an alias to an existing index without removing any existing
// alias bindings.
func (h *ReindexHelper) addAlias(ctx context.Context, indexName, aliasName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexName = strings.TrimPrefix(indexName, "/")
	aliasName = strings.TrimPrefix(aliasName, "/")

	action := map[string]any{
		"actions": []map[string]any{
			{"add": map[string]any{"index": indexName, "alias": aliasName}},
		},
	}
	body, _ := json.Marshal(action)

	res, err := h.es.Indices.UpdateAliases(
		bytes.NewReader(body),
		h.es.Indices.UpdateAliases.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("add alias %s to %s: %w", aliasName, indexName, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("add alias %s to %s: %s — %s", aliasName, indexName, res.Status(), string(respBody))
	}
	return nil
}
