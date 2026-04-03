package migrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
)

// migrationsIndexMapping is the ES index mapping for the .migrations tracking index.
// Document _id equals the migration version string for idempotent upserts.
const migrationsIndexMapping = `{
	"mappings": {
		"properties": {
			"version":      { "type": "keyword" },
			"description": { "type": "text" },
			"applied_at":   { "type": "date" }
		}
	}
}`

// AppliedStore manages the .migrations ES index used to track which migration
// versions have been applied. It is safe for concurrent use across multiple
// application instances — all operations are cluster-scoped via the shared ES
// index.
//
// Each applied migration is stored as a document with _id equal to its version
// string, making Record idempotent (upsert semantics) and Remove safe to call
// even if the document no longer exists.
type AppliedStore struct {
	es    *elasticsearch.Client
	index string
	lg    logger.Logger
}

// NewAppliedStore constructs an AppliedStore that targets the given ES index name.
// If index is empty, ".migrations" is used.
func NewAppliedStore(es *elasticsearch.Client, index string, lg logger.Logger) *AppliedStore {
	if index == "" {
		index = ".migrations"
	}
	return &AppliedStore{es: es, index: index, lg: lg}
}

// Ensure creates the migrations index if it does not already exist.
// It is idempotent: calling it when the index already exists succeeds with no
// side effects.
func (s *AppliedStore) Ensure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := s.es.Indices.Exists([]string{s.index}, s.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("check migrations index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	// 404 → create the index.
	res, err = s.es.Indices.Create(
		s.index,
		s.es.Indices.Create.WithBody(bytes.NewReader([]byte(migrationsIndexMapping))),
		s.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create migrations index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create migrations index: %s — %s", res.Status(), string(body))
	}

	s.lg.Info("migrations index created", "index", s.index)
	return nil
}

// Applied returns the set of migration versions that have been recorded as
// applied in the migrations index.
func (s *AppliedStore) Applied(ctx context.Context) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := s.es.Search(
		s.es.Search.WithContext(ctx),
		s.es.Search.WithIndex(s.index),
		s.es.Search.WithSize(1000),
	)
	if err != nil {
		return nil, fmt.Errorf("search applied migrations: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			// Index doesn't exist yet — treat as zero applied.
			return map[string]bool{}, nil
		}
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search applied migrations: %s — %s", res.Status(), string(body))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Version string `json:"version"`
				} `json:"_source"` // We only care about _id, so ignore _source.
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode applied migrations response: %w", err)
	}

	applied := make(map[string]bool, len(result.Hits.Hits))
	for _, h := range result.Hits.Hits {
		applied[h.Source.Version] = true
	}
	return applied, nil
}

// Record marks a migration as applied by upserting a document with _id equal to
// the migration version. It is idempotent — calling it multiple times for the same
// version has the same effect as calling it once. This makes it safe to retry
// after a crash.
func (s *AppliedStore) Record(ctx context.Context, m Migration) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	doc := map[string]any{
		"version":     m.Version(),
		"description": m.Description(),
		"applied_at":  time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("encode migration record: %w", err)
	}

	res, err := s.es.Index(
		s.index,
		bytes.NewReader(body),
		s.es.Index.WithContext(ctx),
		s.es.Index.WithDocumentID(m.Version()),
	)
	if err != nil {
		return fmt.Errorf("record migration %s: %w", m.Version, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("record migration %s: %s — %s", m.Version, res.Status(), string(respBody))
	}
	return nil
}

// Remove deletes the migration record for the given version. It returns nil if
// the document does not exist (404), making it safe to call on an already-removed
// migration.
func (s *AppliedStore) Remove(ctx context.Context, version string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := s.es.Delete(s.index, version, s.es.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("remove migration record %s: %w", version, err)
	}
	defer res.Body.Close()
	// 404 means it was already gone — not an error.
	if res.IsError() && res.StatusCode != 404 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("remove migration record %s: %s — %s", version, res.Status(), string(respBody))
	}
	return nil
}
