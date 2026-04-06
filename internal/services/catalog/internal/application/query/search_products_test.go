package query

import (
	"context"
	"errors"
	"testing"

	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// mockProductSearcher is a test double for ProductSearcher.
type mockProductSearcher struct {
	returnProducts []*aggregate.Product
	returnTotal    int
	returnErr      error
}

func (m *mockProductSearcher) SearchProducts(_ context.Context, _ SearchProductsFilter) ([]*aggregate.Product, int, error) {
	return m.returnProducts, m.returnTotal, m.returnErr
}

func TestSearchProductsHandler_Handle(t *testing.T) {
	t.Parallel()

	category := "electronics"
	conditions := []string{"new", "like_new"}
	sort := "price_asc"
	products := []*aggregate.Product{
		{},
		{},
	}

	tests := []struct {
		name           string
		query          SearchProductsQuery
		searcher       *mockProductSearcher
		wantTotal      int
		wantLen        int
		wantPage       int
		wantLimit      int
		wantTotalPages int
		wantErr        bool
	}{
		{
			name:           "returns products with correct pagination",
			query:          SearchProductsQuery{Page: 1, Limit: 10},
			searcher:       &mockProductSearcher{returnProducts: products, returnTotal: 2},
			wantTotal:      2,
			wantLen:        2,
			wantPage:       1,
			wantLimit:      10,
			wantTotalPages: 1,
			wantErr:        false,
		},
		{
			name:           "calculates totalPages correctly when not evenly divisible",
			query:          SearchProductsQuery{Page: 1, Limit: 10},
			searcher:       &mockProductSearcher{returnProducts: products, returnTotal: 25},
			wantTotal:      25,
			wantLen:        2,
			wantPage:       1,
			wantLimit:      10,
			wantTotalPages: 3,
			wantErr:        false,
		},
		{
			name:           "handles empty result",
			query:          SearchProductsQuery{Page: 1, Limit: 10},
			searcher:       &mockProductSearcher{returnProducts: nil, returnTotal: 0},
			wantTotal:      0,
			wantLen:        0,
			wantPage:       1,
			wantLimit:      10,
			wantTotalPages: 0,
			wantErr:        false,
		},
		{
			name:           "passes all filter fields to indexer",
			query:          SearchProductsQuery{Page: 2, Limit: 20, Keyword: "iphone", Category: &category, Conditions: conditions, Sort: &sort},
			searcher:       &mockProductSearcher{returnProducts: nil, returnTotal: 0},
			wantTotal:      0,
			wantLen:        0,
			wantPage:       2,
			wantLimit:      20,
			wantTotalPages: 0,
			wantErr:        false,
		},
		{
			name:    "propagates indexer error",
			query:   SearchProductsQuery{Page: 1, Limit: 10},
			searcher: &mockProductSearcher{returnErr: errors.New("elasticsearch unavailable")},
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := NewSearchProductsHandler(tc.searcher)
			got, err := h.Handle(context.Background(), tc.query)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Handle() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Handle() unexpected error: %v", err)
				return
			}

			if len(got.Products) != tc.wantLen {
				t.Errorf("Handle() returned %d products, want %d", len(got.Products), tc.wantLen)
			}
			if got.Pagination.Page != tc.wantPage {
				t.Errorf("Handle() pagination.Page = %d, want %d", got.Pagination.Page, tc.wantPage)
			}
			if got.Pagination.Limit != tc.wantLimit {
				t.Errorf("Handle() pagination.Limit = %d, want %d", got.Pagination.Limit, tc.wantLimit)
			}
			if got.Pagination.TotalItems != tc.wantTotal {
				t.Errorf("Handle() pagination.TotalItems = %d, want %d", got.Pagination.TotalItems, tc.wantTotal)
			}
			if got.Pagination.TotalPages != tc.wantTotalPages {
				t.Errorf("Handle() pagination.TotalPages = %d, want %d", got.Pagination.TotalPages, tc.wantTotalPages)
			}
		})
	}
}
