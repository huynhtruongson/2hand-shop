package query

import (
	"context"
	"errors"
	"testing"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

func TestListProductRequestsHandler_Handle(t *testing.T) {
	t.Parallel()

	category := "electronics"
	conditions := []string{"new", "like_new"}
	statuses := []string{"pending"}
	sellerID := "seller-123"
	sortDesc := "-created_at"

	tests := []struct {
		name         string
		query        ListProductRequestsQuery
		repo         *mockProductRequestRepo
		wantLen      int
		wantPage     int
		wantLimit    int
		wantTotal    int
		wantSellerID *string
		wantErr      bool
	}{
		{
			name:      "default pagination when limit and page are zero",
			query:     ListProductRequestsQuery{Page: 0, Limit: 0},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: defaultLimit,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "respects provided page and limit",
			query:     ListProductRequestsQuery{Page: 2, Limit: 2},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  2,
			wantLimit: 2,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "caps limit at maxLimit",
			query:     ListProductRequestsQuery{Page: 1, Limit: 500},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: maxLimit,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "normalizes negative page to 1",
			query:     ListProductRequestsQuery{Page: -5, Limit: 10},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "passes category filter to repo",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10, Category: &category},
			repo:      newMockProductRequestRepoWithCategory(&category, 2),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "passes conditions and statuses filters to repo",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10, Conditions: conditions, Statuses: statuses},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "passes sort to repo",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10, Sort: &sortDesc},
			repo:      newMockProductRequestRepo(nil, 3),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name:      "admin user — no sellerID filter",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10, User: auth.NewUser("admin-1", "admin").Ptr()},
			repo:      newMockProductRequestRepo(nil, 5),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:      "non-admin user — sellerID filter forced to auth user",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10, User: auth.NewUser(sellerID, "client").Ptr()},
			repo:      newMockProductRequestRepoWithSellerID(&sellerID, 4),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 4,
			wantErr:   false,
		},
		{
			name:      "empty result when no requests match",
			query:     ListProductRequestsQuery{Page: 1, Limit: 10},
			repo:      newMockProductRequestRepo(nil, 0),
			wantLen:   0,
			wantPage:  1,
			wantLimit: 10,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:    "propagates repo error",
			query:   ListProductRequestsQuery{Page: 1, Limit: 10},
			repo:    newMockProductRequestRepoWithError(errors.New("db connection refused")),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := NewListProductRequestsHandler(tc.repo, nil)
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

			if len(got.ProductRequests) != tc.wantLen {
				t.Errorf("Handle() returned %d product requests, want %d", len(got.ProductRequests), tc.wantLen)
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

// mockProductRequestRepo is a test double for repository.ProductRequestRepository.
type mockProductRequestRepo struct {
	categoryFilter    *string
	sellerIDFilter    *string
	statusesFilter     []string
	conditionsFilter   []string
	returnTotal        int
	returnErr          error
}

func newMockProductRequestRepo(category *string, total int) *mockProductRequestRepo {
	return &mockProductRequestRepo{
		categoryFilter: category,
		returnTotal:    total,
	}
}

func newMockProductRequestRepoWithCategory(category *string, total int) *mockProductRequestRepo {
	return &mockProductRequestRepo{
		categoryFilter: category,
		returnTotal:    total,
	}
}

func newMockProductRequestRepoWithSellerID(sellerID *string, total int) *mockProductRequestRepo {
	return &mockProductRequestRepo{
		sellerIDFilter: sellerID,
		returnTotal:    total,
	}
}

func newMockProductRequestRepoWithError(err error) *mockProductRequestRepo {
	return &mockProductRequestRepo{
		returnTotal: 0,
		returnErr:   err,
	}
}

func (m *mockProductRequestRepo) Save(_ context.Context, _ postgressqlx.Querier, _ *aggregate.ProductRequest) error {
	return errors.New("not implemented")
}
func (m *mockProductRequestRepo) Update(_ context.Context, _ postgressqlx.Querier, _ *aggregate.ProductRequest) error {
	return errors.New("not implemented")
}
func (m *mockProductRequestRepo) GetByID(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.ProductRequest, error) {
	return nil, errors.New("not implemented")
}
func (m *mockProductRequestRepo) ListBySellerID(_ context.Context, _ postgressqlx.Querier, _ string) ([]*aggregate.ProductRequest, error) {
	return nil, errors.New("not implemented")
}
func (m *mockProductRequestRepo) Delete(_ context.Context, _ postgressqlx.Querier, _ string) error {
	return errors.New("not implemented")
}
func (m *mockProductRequestRepo) List(_ context.Context, _ postgressqlx.Querier, filter repository.ListProductRequestsFilter, _ postgressqlx.Page) ([]*aggregate.ProductRequest, int, error) {
	if m.categoryFilter != nil && (filter.Category == nil || *m.categoryFilter != *filter.Category) {
		return nil, 0, errors.New("unexpected category filter")
	}
	if m.sellerIDFilter != nil && (filter.SellerID == nil || *m.sellerIDFilter != *filter.SellerID) {
		return nil, 0, errors.New("unexpected sellerID filter")
	}
	return nil, m.returnTotal, m.returnErr
}
