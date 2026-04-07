package postgres

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

type OrderRepo struct{}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

var _ repository.OrderRepository = (*OrderRepo)(nil)

func (r *OrderRepo) Save(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	panic("not implemented")
}

func (r *OrderRepo) Update(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	panic("not implemented")
}

func (r *OrderRepo) GetByID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Order, error) {
	panic("not implemented")
}

func (r *OrderRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListOrdersFilter, page postgressqlx.Page) ([]aggregate.Order, int, error) {
	panic("not implemented")
}
