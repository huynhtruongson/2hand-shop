package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/http/types"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// ListProductRequestsHandler is the CQRS query handler for listing product requests.
type ListProductRequestsHandler cqrs.QueryHandler[ListProductRequestsQuery, ListProductRequestsResponse]

// ListProductRequestsQuery carries the auth context and filter parameters for listing product requests.
type ListProductRequestsQuery struct {
	User       *auth.User
	Page       int
	Limit      int
	Category   *string
	Conditions []string
	Statuses   []string
	Sort       *string
}

// ListProductRequestsResponse holds the paginated product request listing.
type ListProductRequestsResponse struct {
	ProductRequests []*aggregate.ProductRequest
	Pagination      types.Pagination
}

// listProductRequestsHandler implements ListProductRequestsHandler.
type listProductRequestsHandler struct {
	db   postgressqlx.DB
	repo repository.ProductRequestRepository
}

// NewListProductRequestsHandler returns a ListProductRequestsHandler that uses repo for data retrieval.
func NewListProductRequestsHandler(repo repository.ProductRequestRepository, db postgressqlx.DB) ListProductRequestsHandler {
	return &listProductRequestsHandler{repo: repo, db: db}
}

// Handle processes ListProductRequestsQuery and returns a paginated listing.
//
// Access rules:
//   - Admin (role == "admin"): no sellerID filter applied — sees all requests.
//   - Client/Seller (role == "client"): sellerID is forced to authUser.UserID() — sees only own requests.
func (h *listProductRequestsHandler) Handle(ctx context.Context, q ListProductRequestsQuery) (ListProductRequestsResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = defaultLimit
	}
	if q.Limit > maxLimit {
		q.Limit = maxLimit
	}

	offset := (q.Page - 1) * q.Limit

	filter := repository.ListProductRequestsFilter{
		Category:   q.Category,
		Statuses:   q.Statuses,
		Conditions: q.Conditions,
		Sort:       q.Sort,
	}

	if q.User != nil && !q.User.IsAdmin() {
		sellerID := q.User.UserID()
		filter.SellerID = &sellerID
	}

	productRequests, total, err := h.repo.List(ctx, h.db, filter, postgressqlx.NewPage(q.Limit, offset, 100))
	if err != nil {
		return ListProductRequestsResponse{}, err
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return ListProductRequestsResponse{
		ProductRequests: productRequests,
		Pagination: types.Pagination{
			Page:       q.Page,
			Limit:      q.Limit,
			TotalPages: totalPages,
			TotalItems: total,
		},
	}, nil
}
