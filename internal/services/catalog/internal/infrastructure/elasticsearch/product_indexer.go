package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductIndexer implements the Elasticsearch write-side projection for products.
type ProductIndexer struct {
	client    *Client
	indexName string
	logger    logger.Logger
}

// NewProductIndexer constructs a ProductIndexer.
func NewProductIndexer(client *Client, indexName string, lg logger.Logger) *ProductIndexer {
	return &ProductIndexer{
		client:    client,
		indexName: indexName,
		logger:    lg,
	}
}

// IndexProduct indexes (or replaces) a product document.
func (i *ProductIndexer) IndexProduct(ctx context.Context, p *aggregate.Product, categoryName string) error {
	return i.index(ctx, p, categoryName)
}

// UpdateProduct updates a product document (alias for IndexProduct — ES does a full replace on same ID).
func (i *ProductIndexer) UpdateProduct(ctx context.Context, p *aggregate.Product, categoryName string) error {
	return i.index(ctx, p, categoryName)
}

// DeleteProduct removes a product document by its ID.
func (i *ProductIndexer) DeleteProduct(ctx context.Context, productID string) error {
	if i.client == nil {
		return nil
	}

	res, err := i.client.Elasticsearch().Delete(
		i.indexName,
		productID,
		i.client.Elasticsearch().Delete.WithContext(ctx),
	)
	if err != nil {
		i.logger.Error("elasticsearch delete failed", "error", err, "product_id", productID)
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		i.logger.Error("elasticsearch delete failed", "status", res.Status(), "product_id", productID)
		return nil
	}

	return nil
}

func (i *ProductIndexer) index(ctx context.Context, p *aggregate.Product, categoryName string) error {
	if i.client == nil {
		return nil
	}

	doc := toProductDoc(p, categoryName)
	body, err := json.Marshal(doc)
	if err != nil {
		i.logger.Error("failed to marshal product doc for elasticsearch", "error", err, "product_id", p.ID())
		return nil
	}

	res, err := i.client.Elasticsearch().Index(
		i.indexName,
		bytes.NewReader(body),
		i.client.Elasticsearch().Index.WithDocumentID(p.ID()),
		i.client.Elasticsearch().Index.WithContext(ctx),
		i.client.Elasticsearch().Index.WithRefresh("false"),
	)
	if err != nil {
		i.logger.Error("elasticsearch index failed", "error", err, "product_id", p.ID())
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		i.logger.Error("elasticsearch index failed", "status", res.Status(), "product_id", p.ID())
		return nil
	}

	return nil
}

// String implements fmt.Stringer so that nil clients produce clean output in logs.
func (i *ProductIndexer) String() string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ProductIndexer{index=%s}", i.indexName)
}
