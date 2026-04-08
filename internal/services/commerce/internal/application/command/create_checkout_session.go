package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// CreateCheckoutSessionCommand creates a pending Order and returns a Stripe Checkout Session URL.
type CreateCheckoutSessionCommand struct {
	UserID          string
	SuccessURL      string
	CancelURL       string
	ShippingAddress *customtypes.Address
}

// CreateCheckoutSessionResponse is the result of CreateCheckoutSessionHandler.
type CreateCheckoutSessionResponse struct {
	SessionID string
	URL       string
	OrderID   string
}

type CreateCheckoutSessionHandler struct {
	cartRepo       repository.CartRepository
	orderRepo      repository.OrderRepository
	stripeProvider repository.PaymentProvider
	db             postgressqlx.DB
}

// NewCreateCheckoutSessionHandler returns a new CreateCheckoutSessionHandler.
func NewCreateCheckoutSessionHandler(
	cartRepo repository.CartRepository,
	orderRepo repository.OrderRepository,
	stripeProvider repository.PaymentProvider,
	db postgressqlx.DB,
) *CreateCheckoutSessionHandler {
	return &CreateCheckoutSessionHandler{
		cartRepo:       cartRepo,
		orderRepo:      orderRepo,
		stripeProvider: stripeProvider,
		db:             db,
	}
}

func (h *CreateCheckoutSessionHandler) Handle(ctx context.Context, cmd CreateCheckoutSessionCommand) (*CreateCheckoutSessionResponse, error) {
	var result *CreateCheckoutSessionResponse

	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		cart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			return err
		}

		if cart.ItemCount() == 0 {
			return carterrors.ErrCartEmpty
		}

		orderID := uuid.New().String()

		orderItems := make([]entity.OrderItem, 0, cart.ItemCount())
		for _, ci := range cart.Items() {
			orderItems = append(orderItems, entity.NewOrderItem(
				uuid.New().String(),
				orderID,
				ci.ProductID(),
				ci.ProductName(),
				ci.Price(),
				ci.Currency(),
			))
		}

		order, err := aggregate.NewOrder(
			orderID,
			cmd.UserID,
			orderItems,
			valueobject.CurrencyUSD,
			cmd.ShippingAddress,
		)
		if err != nil {
			return err
		}

		if err := h.orderRepo.Save(ctx, tx, order); err != nil {
			return err
		}

		stripeResult, err := h.stripeProvider.CreateSession(ctx, repository.CreateSessionParams{
			Order:      order,
			UserID:     cmd.UserID,
			SuccessURL: cmd.SuccessURL,
			CancelURL:  cmd.CancelURL,
			Currency:   valueobject.CurrencyUSD.String(),
		})
		if err != nil {
			return err
		}

		result = &CreateCheckoutSessionResponse{
			SessionID: stripeResult.SessionID,
			URL:       stripeResult.URL,
			OrderID:   orderID,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
