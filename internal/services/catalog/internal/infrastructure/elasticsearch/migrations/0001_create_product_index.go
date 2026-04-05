package migrations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/elasticsearch/migrator"
)

// ProductsIndex is the Elasticsearch index name for products.
// Exported so bootstrap can instantiate the ProductIndexer with the same name.
const ProductsIndex = "products"

// productsMapping defines the Elasticsearch index mapping for the products index.
const productsMapping = `{
	"mappings": {
		"properties": {
			"id":            { "type": "keyword" },
			"category_id":   { "type": "keyword" },
			"category_name": { "type": "keyword" },
			"title":		 {"type": "text"},
			"description": 	{ "type": "text" },
			"brand":        { "type": "text" },
			"price":        { "type": "scaled_float", "scaling_factor": 100 },
			"currency":     { "type": "keyword" },
			"condition":    { "type": "keyword" },
			"status":       { "type": "keyword" },
			"images":       { "type": "object","enabled": false },
			"created_at":   { "type": "date" },
			"updated_at":   { "type": "date" }
		}
	}
}`

// Ensure createProductIndexMigration satisfies the Migration interface.
var _ migrator.Migration = (*createProductIndexMigration)(nil)

// createProductIndexMigration creates the products Elasticsearch index.
type createProductIndexMigration struct{}

// NewCreateProductIndexMigration returns a new createProductIndexMigration.
func NewCreateProductIndexMigration() *createProductIndexMigration {
	return &createProductIndexMigration{}
}

// Version implements Migration.
func (m *createProductIndexMigration) Version() string {
	return "0001"
}

// Description implements Migration.
func (m *createProductIndexMigration) Description() string {
	return "Create products index"
}

// Up implements Migration. It is idempotent: if the index already exists it returns nil without error.
func (m *createProductIndexMigration) Up(ctx context.Context, es *elasticsearch.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := es.Indices.Exists([]string{ProductsIndex}, es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("check products index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	res, err = es.Indices.Create(
		ProductsIndex,
		es.Indices.Create.WithBody(bytes.NewReader([]byte(productsMapping))),
		es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create products index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create products index: %s — %s", res.Status(), string(body))
	}

	return nil
}

// Down implements Migration.
func (m *createProductIndexMigration) Down(ctx context.Context, es *elasticsearch.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := es.Indices.Delete([]string{ProductsIndex}, es.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("delete products index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete products index: %s — %s", res.Status(), string(body))
	}

	return nil
}
