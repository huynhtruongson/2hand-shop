-- Create carts table: one cart per user
CREATE TABLE carts (
    id         UUID PRIMARY KEY,
    user_id    VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_carts_user_id ON carts(user_id);

-- Create cart_items table: items belong to a cart, unique product per cart
CREATE TABLE cart_items (
    id           UUID PRIMARY KEY,
    cart_id      UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id   VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price        TEXT NOT NULL,      -- customtypes.Price serialized as string
    currency     VARCHAR(10) NOT NULL, -- ISO 4217 code, always USD for now
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(cart_id, product_id)      -- one row per product in a cart
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
