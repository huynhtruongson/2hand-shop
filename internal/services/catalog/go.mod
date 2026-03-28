module github.com/huynhtruongson/2hand-shop/internal/services/catalog

go 1.25.1

replace github.com/huynhtruongson/2hand-shop/internal/pkg => ../../pkg

require github.com/huynhtruongson/2hand-shop/internal/pkg v0.0.0-00010101000000-000000000000

require (
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/lib/pq v1.11.2 // indirect
	github.com/shopspring/decimal v1.4.0
)
