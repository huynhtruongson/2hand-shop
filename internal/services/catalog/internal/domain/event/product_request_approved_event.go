package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductRequestApprovedEvent is published when an admin approves a product request.
type ProductRequestApprovedEvent struct {
	types.BaseEvent
	ProductRequestPayload
}

// NewProductRequestApprovedEvent creates a ProductRequestApprovedEvent from the given product request.
func NewProductRequestApprovedEvent(pr *aggregate.ProductRequest) ProductRequestApprovedEvent {
	return ProductRequestApprovedEvent{
		BaseEvent: types.NewBaseEvent(
			"catalog.product_request.approved",
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
			Images:       pr.Images(),
			ContactInfo:   pr.ContactInfo(),
			CreatedAt:     pr.CreatedAt(),
		},
	}
}
