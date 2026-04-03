# Elasticsearch Migration in Go

## Approaches Overview

| Approach | Best For |
|---|---|
| Reindex API | Mapping changes within same cluster |
| Index Aliases (zero-downtime) | Production, no downtime allowed |
| Snapshot & Restore | Cross-cluster / major version upgrades |
| Rolling Upgrade | Minor version upgrades in-place |
| Logstash / ETL | Complex field transforms, cross-version |

---

## Versioned Migration Pattern (postgres-style)

Mimics `golang-migrate` — versioned `.go` files registered in a slice, a `.migrations` ES index tracks applied versions.

### Flow

1. App starts → runs migrator
2. Migrator reads `.migrations` index to find applied versions
3. Compares against registered migration files
4. Runs pending `Up()` functions in order
5. Marks each as applied in `.migrations` index

### Core Files

**`migrations/migrator.go`**
```go
type Migration struct {
    Version     string
    Description string
    Up          func(ctx context.Context, es *elasticsearch.Client) error
    Down        func(ctx context.Context, es *elasticsearch.Client) error
}

type Migrator struct {
    es         *elasticsearch.Client
    migrations []*Migration
}

func (m *Migrator) Run(ctx context.Context) error {
    // 1. ensure .migrations index exists
    // 2. load applied versions
    // 3. sort by version string
    // 4. run pending Up(), mark applied
}
```

**`migrations/001_create_products.go`** — simple index creation
```go
var Migration001 = &Migration{
    Version:     "001",
    Description: "create products index",
    Up: func(ctx context.Context, es *elasticsearch.Client) error {
        // es.Indices.Create("products", ...)
    },
    Down: func(ctx context.Context, es *elasticsearch.Client) error {
        // es.Indices.Delete(["products"], ...)
    },
}
```

**`migrations/002_add_category_field.go`** — zero-downtime reindex via alias swap
```go
var Migration002 = &Migration{
    Version:     "002",
    Description: "add category field + reindex",
    Up: func(ctx context.Context, es *elasticsearch.Client) error {
        // 1. create products_v2 with new mapping
        // 2. reindex products → products_v2
        // 3. atomic alias swap: products_live → products_v2
    },
}
```

**`main.go`**
```go
migrator := migrations.New(es)
migrator.Register(migrations.Migration001, migrations.Migration002)
migrator.Run(context.Background())
```

### Key Design Decisions

- **Version as sortable string** (`"001"`, `"002"`) — deterministic order
- **Document ID = version** in `.migrations` — idempotent upserts, safe to retry
- **Aliases from day one** — cutover is always just an alias swap

---

## Is This Best Practice?

**Honest answer: no universal standard exists for ES migrations**, but this pattern is solid for most Go apps.

### What's Good
- Versioned and ordered — same mental model as `golang-migrate`
- Idempotent — safe to re-run on crash
- Alias-based cutover — zero-downtime for mapping changes
- No external tooling dependency

### Shortcomings

| Issue | Problem |
|---|---|
| No distributed lock | Multiple pods race on startup — need optimistic lock (ES `version_type=external` or Redis) |
| No rollback safety | `Down()` isn't called automatically on failure; half-written state possible |
| Blocking reindex | `WaitForCompletion=true` blocks app boot for minutes on large indices |
| No CLI | Teams need `migrate up/down/status/create` — startup-only is painful for ops |

### What Teams Actually Do

| Approach | When Used |
|---|---|
| Pre-deploy init container (k8s) | Most common in production — migrations run before pods start |
| `elasticsearch-evolution` (Java) | Closest thing to Flyway for ES on JVM |
| Terraform / OpenTofu for index templates | Stable schemas managed as IaC |
| Feature flags + dual-write | Massive clusters with long transition windows |
| Avoid migrations entirely | ES is additive — just add new fields, never reindex |

### The Most Important Insight

> ES was designed with schema flexibility in mind. You can add new fields to a live index without any migration. Only reindex when changing an existing field's **type or analyzer**.

**Actual best practice:**
1. Use index templates so new indices get correct mappings by default
2. Only reindex when genuinely changing an existing field type
3. Run reindex as a **pre-deploy job**, not at app startup
4. Use aliases from day one
