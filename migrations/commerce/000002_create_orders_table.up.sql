-- Create orders table: one per checkout
CREATE TABLE orders (
    id               UUID PRIMARY KEY,
    user_id          VARCHAR(255) NOT NULL,
    ref_number       VARCHAR(50) NOT NULL UNIQUE,
    subtotal_amount  TEXT NOT NULL,       -- customtypes.Price serialised as string
    total_amount     TEXT NOT NULL,
    currency         VARCHAR(10) NOT NULL, -- ISO 4217 code
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    shipping_address JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_ref_number ON orders(ref_number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Create order_items table: items belong to an order
CREATE TABLE order_items (
    id           UUID PRIMARY KEY,
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price        TEXT NOT NULL,       -- customtypes.Price serialised as string
    currency     VARCHAR(10) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
