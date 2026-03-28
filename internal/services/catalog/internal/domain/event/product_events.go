package event

import (
	"encoding/json"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// ProductEvent is the interface implemented by all Product domain events.
// It enables the application layer to handle heterogeneous event types
// (e.g. for routing or dispatch) without a concrete type switch.
type ProductEvent interface {
	// EventName returns the event's name, used as the RabbitMQ routing key suffix.
	EventName() string
	// Embedded json.Marshaler so the interface is satisfied implicitly
	// by all event types that define MarshalJSON.
	json.Marshaler
}

// ── ProductCreated ────────────────────────────────────────────────────────────

// ProductCreated is raised when a new product aggregate is successfully created.
type ProductCreated struct {
	ProductID    string                      `json:"product_id"`
	SellerID     string                      `json:"seller_id"`
	CategoryID   string                      `json:"category_id"`
	Title        string                      `json:"title"`
	Description  string                      `json:"description"`
	Price        int64                       `json:"price"` // stored as cents to avoid floating-point issues
	Currency     string                      `json:"currency"`
	Condition    valueobject.Condition        `json:"condition"`
	Status       valueobject.ProductStatus    `json:"status"`
	ImageURLs    []string                    `json:"image_urls,omitempty"`
	CreatedAt    time.Time                   `json:"created_at"`
}

// EventName implements ProductEvent.
func (ProductCreated) EventName() string { return "product.created" }

func (e ProductCreated) MarshalJSON() ([]byte, error) {
	type Alias ProductCreated
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductUpdated ───────────────────────────────────────────────────────────

// ProductUpdated is raised when one or more fields of a product are modified.
type ProductUpdated struct {
	ProductID    string                       `json:"product_id"`
	SellerID     string                       `json:"seller_id"`
	Title        string                       `json:"title,omitempty"`
	Description  string                       `json:"description,omitempty"`
	Price        int64                        `json:"price,omitempty"`
	Currency     string                       `json:"currency,omitempty"`
	Condition    *valueobject.Condition       `json:"condition,omitempty"`
	ImageURLs    []string                     `json:"image_urls,omitempty"`
	UpdatedAt    time.Time                    `json:"updated_at"`
}

// EventName implements ProductEvent.
func (ProductUpdated) EventName() string { return "product.updated" }

func (e ProductUpdated) MarshalJSON() ([]byte, error) {
	type Alias ProductUpdated
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductActivated ─────────────────────────────────────────────────────────

// ProductActivated is raised when a draft product transitions to active.
type ProductActivated struct {
	ProductID    string    `json:"product_id"`
	SellerID     string    `json:"seller_id"`
	ActivatedAt  time.Time `json:"activated_at"`
}

// EventName implements ProductEvent.
func (ProductActivated) EventName() string { return "product.activated" }

func (e ProductActivated) MarshalJSON() ([]byte, error) {
	type Alias ProductActivated
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductSold ──────────────────────────────────────────────────────────────

// ProductSold is raised when an active product is purchased.
type ProductSold struct {
	ProductID    string    `json:"product_id"`
	SellerID     string    `json:"seller_id"`
	OrderID      string    `json:"order_id,omitempty"`
	SoldAt       time.Time `json:"sold_at"`
}

// EventName implements ProductEvent.
func (ProductSold) EventName() string { return "product.sold" }

func (e ProductSold) MarshalJSON() ([]byte, error) {
	type Alias ProductSold
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductArchived ──────────────────────────────────────────────────────────

// ProductArchived is raised when an active product is archived by its seller.
type ProductArchived struct {
	ProductID    string    `json:"product_id"`
	SellerID     string    `json:"seller_id"`
	ArchivedAt   time.Time `json:"archived_at"`
}

// EventName implements ProductEvent.
func (ProductArchived) EventName() string { return "product.archived" }

func (e ProductArchived) MarshalJSON() ([]byte, error) {
	type Alias ProductArchived
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}

// ── ProductDeleted ────────────────────────────────────────────────────────────

// ProductDeleted is raised when a product is permanently removed.
type ProductDeleted struct {
	ProductID    string    `json:"product_id"`
	SellerID     string    `json:"seller_id"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// EventName implements ProductEvent.
func (ProductDeleted) EventName() string { return "product.deleted" }

func (e ProductDeleted) MarshalJSON() ([]byte, error) {
	type Alias ProductDeleted
	return json.Marshal(struct {
		Alias
		EventName string `json:"event_name"`
	}{
		Alias:     Alias(e),
		EventName: e.EventName(),
	})
}
