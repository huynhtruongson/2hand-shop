package event

import (
	"encoding/json"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// ProductRequestEvent is the interface implemented by all ProductRequest domain events.
type ProductRequestEvent interface {
	// EventName returns the event's name, used as the RabbitMQ routing key suffix.
	EventName() string
	json.Marshaler
}

// ── ProductRequestCreated ─────────────────────────────────────────────────────

// ProductRequestCreated is raised when a new ProductRequest aggregate is successfully created.
type ProductRequestCreated struct {
	ProductRequestID string                   `json:"product_request_id"`
	SellerID          string                   `json:"seller_id"`
	CategoryID        string                   `json:"category_id"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	Price             int64                    `json:"price"` // stored as cents to avoid floating-point issues
	Currency          string                   `json:"currency"`
	Condition         valueobject.Condition   `json:"condition"`
	ImageURLs         []string                 `json:"image_urls,omitempty"`
	ExpectedPrice     int64                    `json:"expected_price"` // cents; price the seller expects
	ContactInfo       string                   `json:"contact_info"`
	Status            valueobject.ProductRequestStatus `json:"status"`
	CreatedAt         time.Time                `json:"created_at"`
	DeletedAt         *time.Time               `json:"deleted_at,omitempty"`
}

// EventName implements ProductRequestEvent.
func (ProductRequestCreated) EventName() string { return "product_request.created" }

func (e ProductRequestCreated) MarshalJSON() ([]byte, error) {
	type Alias ProductRequestCreated
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductRequestUpdated ────────────────────────────────────────────────────

// ProductRequestUpdated is raised when one or more mutable fields of a ProductRequest are modified
// by the owning seller while the request is still in pending status.
type ProductRequestUpdated struct {
	ProductRequestID string                   `json:"product_request_id"`
	SellerID          string                   `json:"seller_id"`
	Title             string                   `json:"title,omitempty"`
	Description       string                   `json:"description,omitempty"`
	CategoryID        string                   `json:"category_id,omitempty"`
	Price             int64                    `json:"price,omitempty"`
	Currency          string                   `json:"currency,omitempty"`
	Condition         *valueobject.Condition  `json:"condition,omitempty"`
	ImageURLs         []string                 `json:"image_urls,omitempty"`
	ExpectedPrice     int64                    `json:"expected_price,omitempty"`
	ContactInfo       string                   `json:"contact_info,omitempty"`
	UpdatedAt         time.Time                `json:"updated_at"`
	DeletedAt         *time.Time               `json:"deleted_at,omitempty"`
}

// EventName implements ProductRequestEvent.
func (ProductRequestUpdated) EventName() string { return "product_request.updated" }

func (e ProductRequestUpdated) MarshalJSON() ([]byte, error) {
	type Alias ProductRequestUpdated
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}
