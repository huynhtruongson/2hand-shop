package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductRequestUpdatedEvent is published when a seller updates their product request
// while it is still in pending status.
type ProductRequestUpdatedEvent struct {
	types.BaseEvent
	ProductRequestPayload
}

func NewProductRequestUpdatedEvent(pr *aggregate.ProductRequest) ProductRequestUpdatedEvent {
	return ProductRequestUpdatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product_request.updated", "catalog.events", uuid.NewString()),
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