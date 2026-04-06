package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

func (i *ProductIndexer) SearchProducts(ctx context.Context, filter query.SearchProductsFilter) ([]*aggregate.Product, int, error) {
	if i.client == nil {
		return nil, 0, fmt.Errorf("elasticsearch client is nil")
	}

	from := (max(1, filter.Page) - 1) * filter.Limit

	query := buildSearchProductsQuery(filter)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal es search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{i.indexName},
		Body:  bytes.NewReader(body),
		From:  &from,
		Size:  &filter.Limit,
	}

	res, err := req.Do(ctx, i.client.Elasticsearch())
	if err != nil {
		return nil, 0, fmt.Errorf("es search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("es search response error: %s", res.Status())
	}

	var sr searchHitResponse
	if err := json.NewDecoder(res.Body).Decode(&sr); err != nil {
		return nil, 0, fmt.Errorf("decode es search response: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		products = append(products, toAggregate(&h.Source))
	}

	return products, sr.Hits.Total.Value, nil
}

func toAggregate(doc *productDoc) *aggregate.Product {
	price, _ := customtypes.NewPrice(doc.Price)
	currency, _ := valueobject.NewCurrencyFromString(doc.Currency)
	condition, _ := valueobject.NewConditionFromString(doc.Condition)
	status, _ := valueobject.NewProductStatusFromString(doc.Status)

	return aggregate.UnmarshalProductFromDB(
		doc.ID, doc.CategoryID, doc.Title, doc.Description,
		price, currency, condition, status,
		doc.Images, doc.Brand,
		doc.CreatedAt, doc.UpdatedAt, nil,
	)
}

func buildSearchProductsQuery(filter query.SearchProductsFilter) map[string]any {
	must := []map[string]any{}
	filterClauses := []map[string]any{
		{"term": map[string]any{"status": "published"}},
		// {"bool": map[string]any{"must_not": []map[string]any{{"exists": map[string]any{"field": "deleted_at"}}}}},
	}

	if filter.Keyword != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":     filter.Keyword,
				"fields":    []string{"title^3", "description", "brand^2"},
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}

	if filter.Category != nil && *filter.Category != "" {
		filterClauses = append(filterClauses, map[string]any{
			"term": map[string]any{"category_name": *filter.Category},
		})
	}

	if len(filter.Conditions) > 0 {
		filterClauses = append(filterClauses, map[string]any{
			"terms": map[string]any{"condition": filter.Conditions},
		})
	}

	boolQuery := map[string]any{
		"filter": filterClauses,
	}
	if len(must) > 0 {
		boolQuery["must"] = must
	}

	return map[string]any{
		"query": map[string]any{
			"bool": boolQuery,
		},
		"track_total_hits": true,
		"sort":             buildSortClause(filter.Sort, filter.Keyword),
	}
}

func buildSortClause(sort *string, keyword string) []map[string]any {
	sortClauses := buildPrimarySort(sort)
	if keyword != "" {
		sortClauses = append(sortClauses, map[string]any{"_score": map[string]any{"order": "desc"}})
	}
	return sortClauses
}

// buildPrimarySort returns the primary sort clause(s) for the given sort value.
// nil or empty sort returns created_at desc.
func buildPrimarySort(sort *string) []map[string]any {
	if sort == nil || *sort == "" {
		return []map[string]any{{"created_at": map[string]any{"order": "desc"}}}
	}
	switch *sort {
	case "-created_at":
		return []map[string]any{{"created_at": map[string]any{"order": "desc"}}}
	case "-price":
		return []map[string]any{{"price": map[string]any{"order": "desc"}}}
	case "price":
		return []map[string]any{{"price": map[string]any{"order": "asc"}}}
	case "created_at":
		return []map[string]any{{"created_at": map[string]any{"order": "asc"}}}
	default:
		return []map[string]any{{"created_at": map[string]any{"order": "desc"}}}
	}
}

// searchHitResponse is the minimal structure needed to decode the ES _search body.
type searchHitResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source productDoc `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
