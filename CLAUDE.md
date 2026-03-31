# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 📌 Project Overview
A microservice-based second-hand marketplace in **Go (Golang) + React 19**. Four services follow DDD + CQRS with event-driven communication via RabbitMQ:
- **Identity** — User management, Auth/AuthZ via AWS Cognito, Profiles
- **Catalog** — Products, Categories, Item conditions (PostgreSQL + Elasticsearch)
- **Commerce** — Orders, Cart, Payments via Stripe
- **Notification** — Email, Push, In-app real-time

---

## 🧰 Tech Stack

### Backend
- **Language:** Golang
- **Architecture:** Microservices + DDD + CQRS
- **API Gateway:** Tyk
- **Message Broker:** RabbitMQ (via Amazon MQ on AWS)
- **Cloud:** AWS

### Frontend
- **Framework:** React v19
- **Routing:** React Router v7 (file-based routes)
- **Styling:** Tailwind CSS + shadcn/ui

---

## 🛠️ Build, Run & Test Commands

All automation is via [Taskfile](https://taskfile.dev). Run `task --list` to see all tasks.

### Common Tasks

```bash
# Start full local stack (Docker Compose)
task dev:up

# Stop local stack
task dev:down

# Run a single service locally
task run:<service>       # e.g. task run:catalog

# Apply DB migrations for a service
task migrate:<service>   # e.g. task migrate:identity

# Run all unit tests
task test:unit

# Run BDD/integration tests for a service
task test:bdd:<service>  # e.g. task test:bdd:catalog

# Lint all services
task lint

# Build all service Docker images
task build
```

### Local Stack (Docker Compose)

The local environment lives in `deployments/local/`. It starts:
- PostgreSQL (one instance, multiple databases per service)
- RabbitMQ management UI at `http://localhost:15672` (guest/guest)
- Redis
- Elasticsearch at `http://localhost:9200`
- All four microservices + Tyk gateway

---

## 🗺️ Bounded Contexts (DDD)

| Context | Responsibility |
|---|---|
| **Identity** | User management, Auth/AuthZ via AWS Cognito, Profiles |
| **Catalog** | Products, Categories, Item conditions (PostgreSQL + Elasticsearch) |
| **Commerce** | Orders, Cart, Payments via Stripe |
| **Notification** | Email, Push, In-app real-time |

---

## 📐 CQRS Pattern

```
Command Side                        Query Side
────────────                        ──────────
POST /products                      GET /products?q=...
      │                                    │
┌─────▼──────┐                    ┌────────▼──────┐
│  Command   │                    │ Query Handler │
│  Handler   │                    │               │
└─────┬──────┘                    └────────┬──────┘
      │                                    │
┌─────▼──────┐  Domain Event     ┌─────────▼──────┐
│  Postgres  │──product.created─►│ Elasticsearch  │
│  (Write)   │                   │    (Read)      │
└────────────┘                   └────────────────┘
```

### Naming Conventions

| Artifact | Pattern | Example |
|---|---|---|
| Command struct | `VerbNounCommand` | `CreateProductCommand` |
| Command handler | `VerbNounHandler` | `CreateProductHandler` |
| Query struct | `GetNounQuery` / `ListNounsQuery` | `GetProductQuery` |
| Query handler | `GetNounHandler` / `ListNounsHandler` | `ListProductsHandler` |
| Domain event struct | `NounVerbed` (past tense) | `ProductCreated` |
| Event handler | `On<NounVerbed>Handler` | `OnProductCreatedHandler` |
| Repository interface | `NounRepository` | `ProductRepository` |
| DTO (request) | `VerbNounRequest` | `CreateProductRequest` |
| DTO (response) | `NounResponse` | `ProductResponse` |

---

## 📨 Event-Driven Architecture (EDA)

### RabbitMQ Topology

| Exchange | Type | Purpose |
|---|---|---|
| `identity.events` | topic | User/auth domain events |
| `catalog.events` | topic | Product/category domain events |
| `commerce.events` | topic | Order/payment domain events |
| `notification.events` | topic | Notification triggers |

### Routing Key Format

```
<bounded-context>.<aggregate>.<event-name>

Examples:
  catalog.product.created
  catalog.product.updated
  catalog.product.deleted
  commerce.order.placed
  commerce.order.cancelled
  identity.user.registered
```

### Event Payload Schema

All events share a common envelope. The `payload` field is JSON-serialized domain event data.

```go
type EventMessage struct {
    EventID       string          `json:"event_id"`       // UUID v4
    EventName     string          `json:"event_name"`     // e.g. "product.created"
    AggregateID   string          `json:"aggregate_id"`   // UUID of the root aggregate
    AggregateType string          `json:"aggregate_type"` // e.g. "Product"
    OccurredAt    time.Time       `json:"occurred_at"`
    Payload       json.RawMessage `json:"payload"`
}
```

### Event Flow Example

```
CatalogService (producer)
  └─ product created → publishes to exchange: catalog.events
                         routing key: catalog.product.created
  └─ CatalogService (query handler) reads from Elasticsearch

NotificationService (consumer)
  └─ binds queue: notification.catalog.product.created
       → OnProductCreatedHandler → triggers seller confirmation email

CommerceService (producer)
  └─ order placed → publishes to exchange: commerce.events
                      routing key: commerce.order.placed

NotificationService (consumer)
  └─ binds queue: notification.commerce.order.placed
       → OnOrderPlacedHandler → sends order confirmation to buyer
```

### Dead Letter Policy
All queues bind a dead-letter exchange (`<queue-name>.dlx`). Messages are retried 3 times before being routed to the DLX for manual inspection.

---

## 🌐 API Gateway (Tyk)

Gateway config lives in `gateway/`. All external traffic enters through Tyk.

### Route Convention

```
/api/v1/<bounded-context>/<resource>
```

Examples:
```
/api/v1/catalog/products        → Catalog service
/api/v1/identity/users          → Identity service
/api/v1/commerce/orders         → Commerce service
```

---

## 🗄️ Database Per Service

| Service | Write DB | Read / Cache |
|---|---|---|
| Identity | PostgreSQL | Redis (sessions) |
| Catalog | PostgreSQL | Elasticsearch |
| Commerce | PostgreSQL | Redis (cart) |
| Notification | PostgreSQL | — |

Migration files live in `migrations/<bounded-context>/`. Run via `task migrate:<service>`.

---

## ❗ Error Handling Conventions

All errors flow through `AppError` in `internal/pkg/errors/`. It is constructed with `NewAppError(kind, code, message)` and extended with builder methods:

```go
NewAppError(KindNotFound, "PRODUCT_NOT_FOUND", "product not found").
    WithMeta("product_id", id).
    WithCause(err)
```

### Kind → HTTP Status

`Kind` is the coarse classification that drives HTTP status. Key kinds:

| Kind | HTTP |
|---|---|
| `KindNotFound` | 404 |
| `KindConflict` | 409 |
| `KindValidation` |400|
| `KindBadRequest` | 400 |
| `KindUnauthorized` | 401 |
| `KindForbidden` | 403 |
| `KindInternal` | 500 |
| `KindRateLimit` | 429 |

### Error Response Body

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "product not found",
    "details": {
      "product_id": "required field"
    }
  }
}
```

The `details` field is only present for validation errors.

---

## 🏛️ Domain Rules per Bounded Context

### Identity
- A `User` aggregate must have a unique email address (enforced at DB level and domain layer).
- Auth/AuthZ is delegated to **AWS Cognito** — user credentials and tokens are managed by Cognito, not by this service.
- This service manages user profile data (display name, avatar, preferences) linked to the Cognito `sub` claim.
- Session tokens are stored in Redis with a TTL; expiry is enforced in middleware.

### Catalog
- A `Product` aggregate must have: non-empty `Title`, a valid `CategoryID`, a positive `Price`, and a valid `Condition` (`new`, `like_new`, `good`, `fair`, `poor`).
- Status transitions: `draft` → `published` → `sold` | `archived`. No backward transitions.

### Commerce
- An `Order` can only be created from an `published` product.
- Once an `Order` is `placed`, the product status moves to `sold` (via domain event).
- Payment must be confirmed before the order moves to `confirmed`.
- Order status flow: `pending` → `confirmed` → `shipped` → `delivered` | `cancelled`.

### Notification
- Notification service is purely reactive — it only consumes events, it never calls other services directly.
- Delivery channels: `email`, `push`, `in_app`. Each notification type declares its supported channels.

---

## 🔗 Inter-Service Communication

### Async (preferred) — RabbitMQ
Used for all cross-domain state propagation and side effects. See EDA section above.

### Sync — gRPC
Used only when a service needs an immediate response from another service during a request lifecycle.

| Consumer | Provider | Use Case |
|---|---|---|
| Commerce | Identity (gRPC) | Validate user exists before placing order |
| Commerce | Catalog (gRPC) | Fetch product price/status at order time |
| Notification | Identity (gRPC) | Resolve user contact details (email, push token) |

Proto files live in `internal/services/<service>/internal/transports/grpc/proto/`.

**Rule:** If eventual consistency is acceptable, use async events. Use gRPC only when you need a synchronous guarantee within the same request.

---

## 🎨 Code Style & Patterns

### Dependency Injection
Constructor injection everywhere. No global state, no `init()` side effects. Services and handlers are wired in `internal/bootstrap/`.

```go
// Correct
func NewCreateProductHandler(repo domain.ProductRepository, publisher port.EventPublisher) *CreateProductHandler {
    return &CreateProductHandler{repo: repo, publisher: publisher}
}
```

### Linting
Uses `golangci-lint`. Config at repo root (`.golangci.yml`). Run via `task lint`. CI blocks on lint failures.

Key enabled linters: `errcheck`, `govet`, `staticcheck`, `revive`, `exhaustive`, `wrapcheck`.

### Package Layout Rules
- `domain/` has **zero** external imports — no infrastructure, no framework packages.
- `application/` imports `domain/` only.
- `infrastructure/` imports `application/` and external packages.
- `transports/` imports `application/` (DTOs, commands/queries) only — no direct domain imports.

### General Conventions
- All exported functions and types must have godoc comments.
- Use `context.Context` as the first argument in all service and handler methods.
- Table-driven tests for unit tests (`_test.go` files alongside the package).
- BDD tests (Gherkin `.feature` files) live in `test/features/` per service.

---

## 🗂️ Project Folder Structure

```
2hand-shop/
│
├── CLAUDE.md                    # Claude Code guidance
├── README.md                    # Project overview
├── Taskfile.yml                 # Task automation (dev/test/build scripts)
│
├── deployments/                 # IaC & deployment configs
│   └── local/                   #   Local dev setup via Docker Compose
│
├── docs/                       # Architecture & design docs
│
├── gateway/                    # Tyk API Gateway config & API definitions
│
├── internal/                   # 🔒 All application code — not importable by other modules
│   ├── pkg/                    # Shared libraries reused across all services
│   │   ├── cqrs/               #   Base command & query interfaces
│   │   ├── customtypes/        #   Shared custom types (e.g. file attachments)
│   │   ├── errors/             #   App error types & wrappers
│   │   ├── http/               #   HTTP helpers, pagination & response types
│   │   ├── logger/             #   Logger interface + zerolog implementation
│   │   ├── middleware/         #   HTTP middleware (auth, logging, request ID)
│   │   ├── migration/          #   Generic Postgres migration runner
│   │   ├── postgressqlx/       #   sqlx wrappers (pagination, transactions)
│   │   ├── rabbitmq/           #   RabbitMQ client (connection, producer, consumer, manager)
│   │   └── utils/              #   Shared utility functions
│   │
│   └── services/               # All microservice implementations
│       └── {service_name}/     # Each service follows Clean Architecture + DDD
│           ├── cmd/            #   Entry points (app, migration)
│           ├── config/         #   Configuration loaders (env, Postgres, HTTP, etc.)
│           ├── internal/
│           │   ├── domain/     #   🔴 DDD Core — zero external dependencies
│           │   │   ├── aggregate/
│           │   │   ├── entity/
│           │   │   ├── valueobject/
│           │   │   ├── event/  #   Domain events (published via RabbitMQ)
│           │   │   ├── repository/  #   Repository interfaces
│           │   │   └── service/     #   Domain services
│           │   ├── application/  #   🟡 Use-case layer (CQRS)
│           │   │   ├── command/     #   🟣 Write side (Create, Update, Delete)
│           │   │   ├── query/      #   🟣 Read side (Get, List)
│           │   │   ├── eventhandler/  #  Handles incoming RabbitMQ events
│           │   │   ├── dto/         #   Request / response DTOs
│           │   │   └── port/        #   Port interfaces (event publisher, read model)
│           │   ├── infrastructure/  #   🟢 Adapters — external dependencies
│           │   │   ├── persistence/  #   Postgres repositories
│           │   │   ├── messaging/    #   RabbitMQ producer / consumer
│           │   │   └── server/       #   HTTP / gRPC server
│           │   ├── transports/  #   🟠 Delivery layer — HTTP handlers, gRPC, WebSocket
│           │   │   ├── http/
│           │   │   └── grpc/
│           │   └── bootstrap/  #   DI wiring & app startup
│           ├── test/           #   BDD tests (features, steps, fixtures)
│           ├── Dockerfile
│           └── go.mod
│
├── migrations/                 # SQL migration files, per bounded context
│
├── scripts/                    # Dev helper scripts (DB init, etc.)
│
└── web/                       # 🟢 Frontend (React 19) — not yet scaffolded
```