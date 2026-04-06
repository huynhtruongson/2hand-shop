package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductRequestRejectedEvent is published when an admin rejects a product request.
type ProductRequestRejectedEvent struct {
	types.BaseEvent
	ProductRequestPayload
}

// NewProductRequestRejectedEvent creates a ProductRequestRejectedEvent from the given product request.
func NewProductRequestRejectedEvent(pr *aggregate.ProductRequest) ProductRequestRejectedEvent {
	return ProductRequestRejectedEvent{
		BaseEvent: types.NewBaseEvent(
			"catalog.product_request.rejected",
			"catalog.events",
			uuid.NewString(),
		),
		ProductRequestPayload: ProductRequestPayload{
			ID:            pr.ID(),
			SellerID:      pr.SellerID(),
			CategoryID:    pr.CategoryID(),
			Title:         pr.Title(),
			Description:   pr.Description(),
			Brand:         pr.Brand(),
			ExpectedPrice: pr.ExpectedPrice().String(),
			Currency:      pr.Currency().String(),
			Condition:     pr.Condition().String(),
			Status:        pr.Status().String(),
			Images:        pr.Images(),
			ContactInfo:   pr.ContactInfo(),
			CreatedAt:     pr.CreatedAt(),
		},
	}
}
