package query

import (
	"context"
	"errors"
	"testing"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

func TestListProductHandler_Handle(t *testing.T) {
	t.Parallel()

	category := "electronics"

	tests := []struct {
		name      string
		query     ListProductQuery
		repo      *mockProductReadRepo
		wantLen   int
		wantPage  int
		wantLimit int
		wantTotal int
		wantErr   bool
	}{
		{
			name:      "default pagination when limit and page are zero",
			query:     ListProductQuery{Page: 0, Limit: 0},
			repo:      newMockProductReadRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: defaultLimit,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "respects provided page and limit",
			query:     ListProductQuery{Page: 2, Limit: 2},
			repo:      newMockProductReadRepo(nil, 3),
			wantLen:   0,
			wantPage:  2,
			wantLimit: 2,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "caps limit at maxLimit",
			query:     ListProductQuery{Page: 1, Limit: 500},
			repo:      newMockProductReadRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: maxLimit,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "normalizes negative page to 1",
			query:     ListProductQuery{Page: -5, Limit: 10},
			repo:      newMockProductReadRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "passes category filter to repo",
			query:     ListProductQuery{Page: 1, Limit: 10, Category: &category},
			repo:      newMockProductReadRepo(&category, 2),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "empty result when no products match",
			query:     ListProductQuery{Page: 1, Limit: 10},
			repo:      newMockProductReadRepo(nil, 0),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:    "propagates repo error",
			query:   ListProductQuery{Page: 1, Limit: 10},
			repo:    newMockProductReadRepoWithError(errors.New("db connection refused")),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := NewListProductHandler(tc.repo, nil)
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

			wantTotalPages := (tc.wantTotal + tc.wantLimit - 1) / tc.wantLimit
			if got.Pagination.TotalPages != wantTotalPages {
				t.Errorf("Handle() pagination.TotalPages = %d, want %d", got.Pagination.TotalPages, wantTotalPages)
			}
		})
	}
}

// mockProductReadRepo is a test double for port.ProductReadRepository.
type mockProductReadRepo struct {
	categoryFilter *string
	returnTotal    int
	returnErr      error
}

func newMockProductReadRepo(category *string, total int) *mockProductReadRepo {
	return &mockProductReadRepo{
		categoryFilter: category,
		returnTotal:    total,
	}
}

func newMockProductReadRepoWithError(err error) *mockProductReadRepo {
	return &mockProductReadRepo{
		returnTotal: 0,
		returnErr:   err,
	}
}

func (m *mockProductReadRepo) Save(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Product) error {
	return errors.New("not implemented")
}
func (m *mockProductReadRepo) Update(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Product) error {
	return errors.New("not implemented")
}
func (m *mockProductReadRepo) Delete(_ context.Context, _ postgressqlx.Querier, _ string) error {
	return errors.New("not implemented")
}
func (m *mockProductReadRepo) GetByID(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Product, error) {
	return nil, errors.New("not implemented")
}
func (m *mockProductReadRepo) List(_ context.Context, _ postgressqlx.Querier, filter repository.ListProductsFilter, _ postgressqlx.Page) ([]aggregate.Product, int, error) {
	if m.categoryFilter != nil && filter.Category != nil && *m.categoryFilter != *filter.Category {
		return nil, 0, errors.New("unexpected category filter")
	}
	return nil, m.returnTotal, m.returnErr
}
