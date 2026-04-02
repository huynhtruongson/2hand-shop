-- migrate up
CREATE TABLE products (
    id TEXT NOT NULL,
    category_id TEXT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    brand VARCHAR(100),
    price VARCHAR(100) NOT NULL,
    currency VARCHAR(50) NOT NULL,
    condition VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    images JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT products_pkey PRIMARY KEY (id),
    CONSTRAINT products_category_id_fkey 
        FOREIGN KEY (category_id) 
        REFERENCES categories (id) 
        ON UPDATE CASCADE 
        ON DELETE RESTRICT
);

CREATE INDEX idx_products_category_id ON products (category_id);
CREATE INDEX idx_products_status ON products (status);
CREATE INDEX idx_products_condition ON products (condition);
CREATE INDEX idx_products_deleted_at ON products (deleted_at);