package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

type ProductRequestCreatedEvent struct {
	types.BaseEvent
	ProductRequestPayload
}

type ProductRequestPayload struct {
	ID           string                  `json:"id"`
	SellerID     string                  `json:"seller_id"`
	CategoryID   string                  `json:"category_id"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Brand        *string                 `json:"brand,omitempty"`
	ExpectedPrice string                  `json:"expected_price"`
	Currency     string                  `json:"currency"`
	Condition    string                  `json:"condition"`
	Status       string                  `json:"status"`
	Images       customtypes.Attachments `json:"images,omitempty"`
	ContactInfo  string                  `json:"contact_info"`
	CreatedAt    time.Time               `json:"created_at"`
}

func NewProductRequestCreatedEvent(pr *aggregate.ProductRequest) ProductRequestCreatedEvent {
	return ProductRequestCreatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product_request.created", "catalog.events", uuid.NewString()),
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
