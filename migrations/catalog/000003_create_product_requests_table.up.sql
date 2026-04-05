CREATE TABLE product_requests (
    id                  TEXT                NOT NULL,
    seller_id           TEXT                NOT NULL,
    category_id         TEXT                NOT NULL,
    title               VARCHAR(255)        NOT NULL,
    description         TEXT                NOT NULL,
    brand               VARCHAR(255)        NULL,
    currency            VARCHAR(10)         NOT NULL,
    condition           VARCHAR(50)         NOT NULL,
    status              VARCHAR(50)         NOT NULL DEFAULT 'pending',
    images              JSONB               NOT NULL DEFAULT '[]',
    expected_price      VARCHAR(100)        NOT NULL,
    contact_info        TEXT                NOT NULL,
    admin_reject_reason TEXT                NULL,
    admin_note          TEXT                NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT pk_product_requests PRIMARY KEY (id),
    CONSTRAINT products_category_id_fkey 
        FOREIGN KEY (category_id) 
        REFERENCES categories (id) 
        ON UPDATE CASCADE 
        ON DELETE RESTRICT
);

-- Indexes
CREATE INDEX idx_product_requests_seller_id   ON product_requests (seller_id);
CREATE INDEX idx_product_requests_category_id ON product_requests (category_id);
CREATE INDEX idx_product_requests_status       ON product_requests (status);
CREATE INDEX idx_product_requests_deleted_at   ON product_requests (deleted_at);
