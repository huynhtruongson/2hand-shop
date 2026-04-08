package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// cartModel mirrors the carts DB table.
type cartModel struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// cartItemModel mirrors the cart_items DB table.
type cartItemModel struct {
	ID          string            `db:"id"`
	CartID      string            `db:"cart_id"`
	ProductID   string            `db:"product_id"`
	ProductName string            `db:"product_name"`
	Price       customtypes.Price `db:"price"`
	Currency    string            `db:"currency"`
	AddedAt     time.Time         `db:"added_at"`
}

// toCartItem reconstructs a domain CartItem from a DB row.
func (m cartItemModel) toCartItem() (entity.CartItem, error) {
	currency, err := valueobject.NewCurrencyFromString(m.Currency)
	if err != nil {
		return entity.CartItem{}, carterrors.ErrInternal.WithCause(err).WithInternal("cartItemModel.toCartItem: parse currency")
	}

	return entity.NewCartItem(m.ID, m.CartID, m.ProductID, m.ProductName, m.Price, currency), nil
}

// toCartItemModels converts a slice of domain CartItems to DB models.
func toCartItemModels(cartID string, items []entity.CartItem) []cartItemModel {
	models := make([]cartItemModel, 0, len(items))
	for _, item := range items {
		models = append(models, cartItemModel{
			ID:          item.ID(),
			CartID:      cartID,
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			Price:       item.Price(),
			Currency:    item.Currency().String(),
			AddedAt:     item.AddedAt(),
		})
	}
	return models
}

// CartRepo implements repository.CartRepository using PostgreSQL.
type CartRepo struct{}

func NewCartRepo() *CartRepo {
	return &CartRepo{}
}

var _ repository.CartRepository = (*CartRepo)(nil)

// GetByUserID retrieves a cart by user ID, loading all its items.
// Returns ErrCartNotFound when no cart exists for the given user.
func (r *CartRepo) GetByUserID(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error) {
	const cartQuery = `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
		LIMIT 1`

	var cart cartModel
	err := q.QueryRowxContext(ctx, cartQuery, userID).StructScan(&cart)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, carterrors.ErrCartNotFound
		}
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.GetByUserID: fetch cart")
	}

	const itemsQuery = `
		SELECT id, cart_id, product_id, product_name, price, currency, added_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY added_at ASC`

	var itemRows []cartItemModel
	if err := q.SelectContext(ctx, &itemRows, itemsQuery, cart.ID); err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.GetByUserID: fetch items")
	}

	items := make([]entity.CartItem, 0, len(itemRows))
	for _, row := range itemRows {
		item, err := row.toCartItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return aggregate.UnmarshalCartFromDB(
		cart.ID, cart.UserID, items, cart.CreatedAt, cart.UpdatedAt,
	), nil
}

// Save persists or updates a cart and replaces all its items.
// The cart is upserted using ON CONFLICT (user_id) so a cart always exists
// for each user. All existing cart_items are deleted and re-inserted to
// correctly reflect the current state of the cart domain model.
func (r *CartRepo) Save(ctx context.Context, q postgressqlx.Querier, cart *aggregate.Cart) error {
	// Upsert the cart row.
	const upsertCart = `
		INSERT INTO carts (id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at`

	_, err := q.ExecContext(ctx, upsertCart,
		cart.ID(), cart.UserID(), cart.CreatedAt(), cart.UpdatedAt())
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.Save: upsert cart")
	}

	// Replace all items: delete existing, then batch-insert current state.
	const deleteItems = `DELETE FROM cart_items WHERE cart_id = $1`
	if _, err := q.ExecContext(ctx, deleteItems, cart.ID()); err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.Save: delete items")
	}

	if len(cart.Items()) == 0 {
		return nil
	}

	models := toCartItemModels(cart.ID(), cart.Items())
	// pq.Array needed for the batch insert.
	ids := make([]string, len(models))
	cartIDs := make([]string, len(models))
	productIDs := make([]string, len(models))
	productNames := make([]string, len(models))
	prices := make([]customtypes.Price, len(models))
	currencies := make([]string, len(models))
	addedAts := make([]time.Time, len(models))
	for i, m := range models {
		ids[i] = m.ID
		cartIDs[i] = m.CartID
		productIDs[i] = m.ProductID
		productNames[i] = m.ProductName
		prices[i] = m.Price
		currencies[i] = m.Currency
		addedAts[i] = m.AddedAt
	}

	const insertItems = `
		INSERT INTO cart_items (id, cart_id, product_id, product_name, price, currency, added_at)
		SELECT * FROM UNNEST($1::uuid[], $2::uuid[], $3::varchar[], $4::varchar[],
		                    $5::text[], $6::varchar[], $7::timestamptz[])`

	_, err = q.ExecContext(ctx, insertItems,
		pq.Array(ids), pq.Array(cartIDs), pq.Array(productIDs),
		pq.Array(productNames), pq.Array(prices), pq.Array(currencies),
		pq.Array(addedAts))
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.Save: insert items")
	}

	return nil
}

// Delete removes a cart and all its items by user ID.
func (r *CartRepo) Delete(ctx context.Context, q postgressqlx.Querier, userID string) error {
	const query = `DELETE FROM carts WHERE user_id = $1`
	_, err := q.ExecContext(ctx, query, userID)
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("CartRepo.Delete")
	}
	return nil
}
