CREATE TABLE categories (
    id          TEXT        PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    icon_url    VARCHAR(255),
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_categories_slug ON categories (slug);
CREATE INDEX idx_categories_deleted_at ON categories (deleted_at);

INSERT INTO categories (id, name, description, slug, sort_order) VALUES
    (gen_random_uuid()::text, 'Clothes', 'All types of second-hand clothing', 'clothes', 1),
    (gen_random_uuid()::text, 'Accessories', 'Bags, watches, jewelry and more', 'accessories', 2),
    (gen_random_uuid()::text, 'Shoes', 'Sneakers, boots, heels and footwear', 'shoes', 3);
