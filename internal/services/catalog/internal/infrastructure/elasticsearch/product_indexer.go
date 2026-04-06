package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type productDoc struct {
	ID           string                  `json:"id"`
	CategoryID   string                  `json:"category_id"`
	CategoryName string                  `json:"category_name"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Brand        *string                 `json:"brand,omitempty"`
	Price        string                  `json:"price"`
	Currency     string                  `json:"currency"`
	Condition    string                  `json:"condition"`
	Status       string                  `json:"status"`
	Images       customtypes.Attachments `json:"images"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

type ProductIndexer struct {
	client    *Client
	indexName string
}

func NewProductIndexer(client *Client, indexName string) *ProductIndexer {
	return &ProductIndexer{
		client:    client,
		indexName: indexName,
	}
}

func (i *ProductIndexer) IndexProduct(ctx context.Context, p event.ProductPayload) error {
	doc := productDoc{
		ID:           p.ID,
		CategoryID:   p.CategoryID,
		CategoryName: p.CategoryName,
		Title:        p.Title,
		Description:  p.Description,
		Brand:        p.Brand,
		Price:        p.Price,
		Currency:     p.Currency,
		Condition:    p.Condition,
		Status:       p.Status,
		Images:       p.Images,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdateAt,
	}

	return i.index(ctx, &doc)
}

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
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil
	}

	return nil
}

func (i *ProductIndexer) index(ctx context.Context, p *productDoc) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := i.client.Elasticsearch().Index(
		i.indexName,
		bytes.NewReader(body),
		i.client.Elasticsearch().Index.WithDocumentID(p.ID),
		i.client.Elasticsearch().Index.WithContext(ctx),
		i.client.Elasticsearch().Index.WithRefresh("false"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return err
	}

	return nil
}
