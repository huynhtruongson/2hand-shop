package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type CheckoutCompleted struct {
	types.BaseEvent

	ProductIDs []string `json:"product_ids"`
}

func NewCheckoutCompleted(productIDs []string) CheckoutCompleted {
	return CheckoutCompleted{
		BaseEvent: types.NewBaseEvent(
			"checkout.completed",
			"commerce.events",
			uuid.NewString(),
		),
		ProductIDs: productIDs,
	}
}
