package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	commercevo "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// orderModel mirrors the orders DB table.
type orderModel struct {
	ID              string          `db:"id"`
	UserID          string          `db:"user_id"`
	RefNumber       string          `db:"ref_number"`
	SubtotalAmount  string          `db:"subtotal_amount"`
	TotalAmount     string          `db:"total_amount"`
	Currency        string          `db:"currency"`
	Status          string          `db:"status"`
	ShippingAddress sql.NullString  `db:"shipping_address"` // JSONB serialised as string
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	DeletedAt       sql.NullTime    `db:"deleted_at"`
}

// orderItemModel mirrors the order_items DB table.
type orderItemModel struct {
	ID          string    `db:"id"`
	OrderID     string    `db:"order_id"`
	ProductID   string    `db:"product_id"`
	ProductName string    `db:"product_name"`
	Price       string    `db:"price"`
	Currency    string    `db:"currency"`
	CreatedAt   time.Time `db:"created_at"`
}

// toOrderItem reconstructs a domain OrderItem from a DB row.
func (m orderItemModel) toOrderItem() (entity.OrderItem, error) {
	var price customtypes.Price
	if err := price.Scan(m.Price); err != nil {
		return entity.OrderItem{}, carterrors.ErrInternal.WithCause(err).WithInternal("orderItemModel.toOrderItem: scan price")
	}

	currency, err := commercevo.NewCurrencyFromString(m.Currency)
	if err != nil {
		return entity.OrderItem{}, carterrors.ErrInternal.WithCause(err).WithInternal("orderItemModel.toOrderItem: parse currency")
	}

	return entity.NewOrderItem(m.ID, m.OrderID, m.ProductID, m.ProductName, price, currency), nil
}

// toOrderItemModels converts a slice of domain OrderItems to DB models.
func toOrderItemModels(orderID string, items []entity.OrderItem) []orderItemModel {
	models := make([]orderItemModel, 0, len(items))
	for _, item := range items {
		models = append(models, orderItemModel{
			ID:          item.ID(),
			OrderID:     orderID,
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			Price:       item.Price().String(),
			Currency:    item.Currency().String(),
			CreatedAt:   item.CreatedAt(),
		})
	}
	return models
}

// OrderRepo implements repository.OrderRepository using PostgreSQL.
type OrderRepo struct{}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

var _ repository.OrderRepository = (*OrderRepo)(nil)

// Save persists a new order and its items to the database.
func (r *OrderRepo) Save(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	const insertOrder = `
		INSERT INTO orders (id, user_id, ref_number, subtotal_amount, total_amount, currency, status, shipping_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var shippingJSON []byte
	if order.ShippingAddress() != nil {
		var err error
		shippingJSON, err = json.Marshal(order.ShippingAddress())
		if err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.Save: marshal shipping address")
		}
	}

	_, err := q.ExecContext(ctx, insertOrder,
		order.ID(), order.UserID(), order.RefNumber(),
		order.SubtotalAmount().String(), order.TotalAmount().String(),
		order.Currency().String(), order.Status().String(),
		shippingJSON, order.CreatedAt(), order.UpdatedAt(),
	)
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.Save: insert order")
	}

	return r.saveItems(ctx, q, order.ID(), order.Items())
}

// saveItems persists order items, deleting existing ones first.
func (r *OrderRepo) saveItems(ctx context.Context, q postgressqlx.Querier, orderID string, items []entity.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	const deleteItems = `DELETE FROM order_items WHERE order_id = $1`
	if _, err := q.ExecContext(ctx, deleteItems, orderID); err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.saveItems: delete existing items")
	}

	models := toOrderItemModels(orderID, items)
	ids := make([]string, len(models))
	orderIDs := make([]string, len(models))
	productIDs := make([]string, len(models))
	productNames := make([]string, len(models))
	prices := make([]string, len(models))
	currencies := make([]string, len(models))
	createdAts := make([]time.Time, len(models))
	for i, m := range models {
		ids[i] = m.ID
		orderIDs[i] = m.OrderID
		productIDs[i] = m.ProductID
		productNames[i] = m.ProductName
		prices[i] = m.Price
		currencies[i] = m.Currency
		createdAts[i] = m.CreatedAt
	}

	const insertItems = `
		INSERT INTO order_items (id, order_id, product_id, product_name, price, currency, created_at)
		SELECT * FROM UNNEST($1::uuid[], $2::uuid[], $3::varchar[], $4::varchar[], $5::text[], $6::varchar[], $7::timestamptz[])`

	_, err := q.ExecContext(ctx, insertItems,
		pq.Array(ids), pq.Array(orderIDs), pq.Array(productIDs),
		pq.Array(productNames), pq.Array(prices), pq.Array(currencies),
		pq.Array(createdAts))
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.saveItems: insert items")
	}
	return nil
}

// Update updates an existing order and replaces all its items.
func (r *OrderRepo) Update(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	const updateOrder = `
		UPDATE orders SET
			status = $2,
			subtotal_amount = $3,
			total_amount = $4,
			currency = $5,
			shipping_address = $6,
			updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL`

	var shippingJSON []byte
	if order.ShippingAddress() != nil {
		var err error
		shippingJSON, err = json.Marshal(order.ShippingAddress())
		if err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.Update: marshal shipping address")
		}
	}

	result, err := q.ExecContext(ctx, updateOrder,
		order.ID(), order.Status().String(),
		order.SubtotalAmount().String(), order.TotalAmount().String(),
		order.Currency().String(), shippingJSON, order.UpdatedAt(),
	)
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.Update: update order")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.Update: rows affected")
	}
	if rows == 0 {
		return carterrors.ErrOrderNotFound
	}

	return r.saveItems(ctx, q, order.ID(), order.Items())
}

// GetByID retrieves an order by its ID, loading all its items.
func (r *OrderRepo) GetByID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Order, error) {
	const orderQuery = `
		SELECT id, user_id, ref_number, subtotal_amount, total_amount, currency, status, shipping_address, created_at, updated_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m orderModel
	err := q.QueryRowxContext(ctx, orderQuery, orderID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, carterrors.ErrOrderNotFound
		}
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.GetByID: fetch order")
	}

	items, err := r.loadItems(ctx, q, orderID)
	if err != nil {
		return nil, err
	}

	return r.modelToOrder(m, items)
}

// loadItems retrieves all items for an order.
func (r *OrderRepo) loadItems(ctx context.Context, q postgressqlx.Querier, orderID string) ([]entity.OrderItem, error) {
	const itemsQuery = `
		SELECT id, order_id, product_id, product_name, price, currency, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC`

	var rows []orderItemModel
	if err := q.SelectContext(ctx, &rows, itemsQuery, orderID); err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.loadItems")
	}

	items := make([]entity.OrderItem, 0, len(rows))
	for _, row := range rows {
		item, err := row.toOrderItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// List returns orders matching the filter with pagination.
func (r *OrderRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListOrdersFilter, page postgressqlx.Page) ([]aggregate.Order, int, error) {
	args := []any{}
	where := "WHERE deleted_at IS NULL"
	argIdx := 1

	if filter.BuyerID != nil {
		where += " AND user_id = $" + itoa(argIdx)
		args = append(args, *filter.BuyerID)
		argIdx++
	}
	if len(filter.Statuses) > 0 {
		where += " AND status = ANY($" + itoa(argIdx) + "::varchar[])"
		args = append(args, pq.Array(filter.Statuses))
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM orders " + where
	var total int
	if err := q.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.List: count")
	}

	selectQuery := `
		SELECT id, user_id, ref_number, subtotal_amount, total_amount, currency, status, shipping_address, created_at, updated_at
		FROM orders ` + where + " ORDER BY created_at DESC " + page.SQL()
	args = append(args, page.Limit, page.Offset)

	rows, err := q.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.List: select")
	}
	defer rows.Close()

	orders := make([]aggregate.Order, 0)
	for rows.Next() {
		var m orderModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.List: scan row")
		}
		items, err := r.loadItems(ctx, q, m.ID)
		if err != nil {
			return nil, 0, err
		}
		order, err := r.modelToOrder(m, items)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, *order)
	}

	return orders, total, nil
}

// modelToOrder reconstructs a domain Order from a DB model.
func (r *OrderRepo) modelToOrder(m orderModel, items []entity.OrderItem) (*aggregate.Order, error) {
	var subtotal, total customtypes.Price
	if err := subtotal.Scan(m.SubtotalAmount); err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.modelToOrder: scan subtotal")
	}
	if err := total.Scan(m.TotalAmount); err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.modelToOrder: scan total")
	}

	currency, err := commercevo.NewCurrencyFromString(m.Currency)
	if err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.modelToOrder: parse currency")
	}

	status, err := commercevo.NewOrderStatusFromString(m.Status)
	if err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.modelToOrder: parse status")
	}

	var shipping *commercevo.ShippingAddress
	if m.ShippingAddress.Valid && m.ShippingAddress.String != "" {
		var addr commercevo.ShippingAddress
		if err := json.Unmarshal([]byte(m.ShippingAddress.String), &addr); err != nil {
			return nil, carterrors.ErrInternal.WithCause(err).WithInternal("OrderRepo.modelToOrder: unmarshal shipping address")
		}
		shipping = &addr
	}

	return aggregate.UnmarshalOrderFromDB(
		m.ID, m.UserID, items, m.RefNumber,
		subtotal, total, currency, status, shipping,
		m.CreatedAt, m.UpdatedAt,
	), nil
}

// itoa converts an int to a string without importing strconv.
func itoa(i int) string {
	if i == 1 {
		return "1"
	}
	return string(rune('0'+i)) // simple 1-9, fine for pagination args
}
