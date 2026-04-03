package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
)

// Client wraps the official elastic/go-elasticsearch client.
// It handles connection and ping; index lifecycle is owned by per-index types (e.g. ProductIndex).
type Client struct {
	es    *elasticsearch.Client
	addrs []string
	lg    logger.Logger
}

// NewClient connects to Elasticsearch and pings it.
// Returns (nil, error) on any failure — non-fatal; the caller logs the warning
// and the service starts without the search index.
func NewClient(cfg config.ElasticsearchConfig, lg logger.Logger) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.Address},
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		lg.Warn("failed to create elasticsearch client, running without search index", "error", err)
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	res, err := es.Ping()
	if err != nil {
		lg.Warn("elasticsearch unreachable, running without search index", "error", err)
		return nil, fmt.Errorf("elasticsearch ping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		lg.Warn("elasticsearch ping failed, running without search index", "status", res.Status())
		return nil, fmt.Errorf("elasticsearch ping: status %s", res.Status())
	}

	lg.Info("connected to elasticsearch", "addresses", cfg.Address)
	return &Client{es: es, addrs: []string{cfg.Address}, lg: lg}, nil
}

// Elasticsearch returns the underlying elastic/go-elasticsearch client.
// Indexers use this for document-level operations (Index, Delete, Search, etc.).
func (c *Client) Elasticsearch() *elasticsearch.Client {
	return c.es
}
