package aggregate

import (
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
)

type Cart struct {
	id        string
	userID    string
	items     []entity.CartItem
	createdAt time.Time
	updatedAt time.Time
}

func NewCart(id, userID string, items []entity.CartItem) (*Cart, error) {
	c := &Cart{
		id:        id,
		userID:    userID,
		items:     items,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (c *Cart) ID() string               { return c.id }
func (c *Cart) UserID() string           { return c.userID }
func (c *Cart) Items() []entity.CartItem { return c.items }
func (c *Cart) CreatedAt() time.Time     { return c.createdAt }
func (c *Cart) UpdatedAt() time.Time     { return c.updatedAt }

// ItemCount returns the number of items in the cart.
func (c *Cart) ItemCount() int { return len(c.items) }

// TotalAmount returns the sum of all item prices in the cart.
func (c *Cart) TotalAmount() customtypes.Price {
	var total customtypes.Price
	for _, item := range c.items {
		total = total.Add(item.Price())
	}
	return total
}

// ── Mutations ───────────────────────────────────────────────────────────────

// AddItem adds an item to the cart. If the product already exists, it updates quantity.
func (c *Cart) AddItem(item entity.CartItem) {
	for i, existing := range c.items {
		if existing.ProductID() == item.ProductID() {
			c.items[i] = item
			return
		}
	}
	c.items = append(c.items, item)
	c.updatedAt = time.Now().UTC()
}

// RemoveItem removes an item from the cart by product ID.
func (c *Cart) RemoveItem(productID string) bool {
	for i, item := range c.items {
		if item.ProductID() == productID {
			c.items = append(c.items[:i], c.items[i+1:]...)
			c.updatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

// Clear removes all items from the cart.
func (c *Cart) Clear() {
	c.items = nil
	c.updatedAt = time.Now().UTC()
}

// UnmarshalCartFromDB reconstructs a Cart from persisted data, skipping validation.
func UnmarshalCartFromDB(
	id, userID string,
	items []entity.CartItem,
	createdAt, updatedAt time.Time,
) *Cart {
	return &Cart{
		id:        id,
		userID:    userID,
		items:     items,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (c *Cart) validate() error {
	switch {
	case strings.TrimSpace(c.id) == "":
		return errors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(c.userID) == "":
		return errors.ErrValidation.WithDetail("user_id", "user_id is empty")
	}
	return nil
}
