package postgres

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

type PaymentRepo struct{}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{}
}

var _ repository.PaymentRepository = (*PaymentRepo)(nil)

func (r *PaymentRepo) Save(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	panic("not implemented")
}

func (r *PaymentRepo) Update(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	panic("not implemented")
}

func (r *PaymentRepo) GetByID(ctx context.Context, q postgressqlx.Querier, paymentID string) (*aggregate.Payment, error) {
	panic("not implemented")
}

func (r *PaymentRepo) GetByOrderID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Payment, error) {
	panic("not implemented")
}

func (r *PaymentRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListPaymentsFilter, page postgressqlx.Page) ([]aggregate.Payment, int, error) {
	panic("not implemented")
}
