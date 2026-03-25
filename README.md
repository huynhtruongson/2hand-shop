# 🛍️ Second-Hand Shop — System Design Context

> **Purpose:** This document captures the full system design conversation for a second-hand shop project. Load this as context at the start of new conversations to continue where we left off.

---

## 📌 Project Overview

A personal second-hand shop platform (marketplace-style) where users can list, browse, buy, and sell used items.

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

### DevOps
- **Environments:** `dev` (auto-deploy on merge to `main`) and `prod` (tag-triggered + manual approval)
- **CI/CD:** GitHub Actions
- **Containerization:** Docker + ECR
- **Orchestration:** ECS (Elastic Container Service)
- **IaC:** Terraform

---

## 🗺️ Bounded Contexts (DDD)

| Context | Responsibility |
|---|---|
| **Identity** | User management, Auth/AuthZ, Profiles |
| **Catalog** | Listings, Categories, Item conditions |
| **Commerce** | Orders, Cart, Payments |
| **Fulfillment** | Shipping, Tracking, Returns |
| **Notification** | Email, Push, In-app real-time |
| **Search** | Full-text search, Filtering, Autocomplete |
| **Moderation** | Item review, Report handling, Fraud detection |
| **Analytics** | Reports, Metrics, Admin dashboard |

---

## 🏗️ Microservices

### 1. Identity Service (`:8081`)
- **Write DB:** PostgreSQL
- **Read/Cache:** Redis (sessions + tokens)
- **Domain Events Published:**
  - `user.registered`
  - `user.profile_updated`
  - `user.suspended`
- **Commands:** `RegisterUser`, `Login`, `UpdateProfile`
- **Queries:** `GetUser`, `GetProfile`

### 2. Catalog Service (`:8082`) — Core Domain
- **Write DB:** PostgreSQL
- **Read DB:** OpenSearch/Elasticsearch (CQRS read side)
- **Storage:** S3 (item images via CloudFront CDN)
- **Domain Events Published:**
  - `listing.created`
  - `listing.published`
  - `listing.sold`
  - `listing.taken_down`
- **Commands:** `CreateListing`, `UpdateListing`, `PublishListing`, `TakeDownListing`
- **Queries:** `GetListing`, `GetListingsBySeller`, `GetFeaturedListings`

### 3. Commerce Service (`:8083`)
- **Write DB:** PostgreSQL
- **Read/Cache:** Redis (cart storage)
- **Payment Gateway:** Stripe
- **Domain Events Published:**
  - `order.created`
  - `order.payment_completed`
  - `order.cancelled`
  - `order.refunded`
- **Commands:** `AddToCart`, `Checkout`, `ProcessPayment`, `CancelOrder`
- **Queries:** `GetOrder`, `GetOrderHistory`

### 4. Search Service (`:8085`)
- **Read DB:** OpenSearch/Elasticsearch
- **Consumes Events:**
  - `listing.published` → index document
  - `listing.sold` → update availability flag
  - `listing.taken_down` → remove from index
- **Queries:** `SearchListings`, `FilterListings`, `Suggest` (autocomplete)

### 5. Notification Service (`:8084`)
- **Infrastructure:** AWS SES (email), AWS SNS (push), WebSocket (in-app real-time)
- **Consumes Events:** `order.*`, `listing.*`, `user.*`

### 6. Fulfillment Service (`:8086`)
- **Write DB:** PostgreSQL
- **Shipping Providers:** GHN, GHTK (Vietnam)
- **Commands:** `CreateShipment`, `UpdateTracking`
- **Queries:** `GetTracking`

### 7. Moderation Service
- **Write DB:** PostgreSQL
- **AI:** AWS Rekognition (image moderation)
- **Commands:** `ReportListing`, `ReviewListing`
- **Queries:** `GetPendingReviews`

---

## 📐 CQRS Pattern

```
Command Side                        Query Side
────────────                        ──────────
POST /listings                      GET /listings?q=...
      │                                    │
┌─────▼──────┐                    ┌────────▼──────┐
│  Command   │                    │ Query Handler │
│  Handler   │                    │               │
└─────┬──────┘                    └────────┬──────┘
      │                                    │
┌─────▼──────┐  Domain Event     ┌─────────▼──────┐
│  Postgres  │──listing.created─►│ Elasticsearch  │
│  (Write)   │                   │    (Read)      │
└────────────┘                   └────────────────┘
```

**Key Pattern:** Commands write to PostgreSQL, publish domain events via RabbitMQ. Event consumers update the read-side (Elasticsearch) asynchronously.

---

## 📨 RabbitMQ Exchange Design

```
Exchange: shop.events  (topic)
Dead Letter Exchange: shop.dlx
Retry Exchange: shop.retry (with TTL backoff)

Routing Key        →  Consumer Queue
──────────────────────────────────────────────────
user.*             →  notification.user.queue
listing.*          →  search.index.queue
                   →  notification.listing.queue
                   →  moderation.review.queue
order.*            →  fulfillment.order.queue
                   →  notification.order.queue
                   →  catalog.inventory.queue
```

---

## ☁️ AWS Infrastructure

| Layer | Service |
|---|---|
| DNS | Route 53 |
| CDN | CloudFront |
| Frontend Hosting | S3 |
| Load Balancing | ALB |
| Compute | ECS (Fargate) |
| Container Registry | ECR |
| Write Database | RDS PostgreSQL (Multi-AZ in prod) |
| Cache | ElastiCache Redis |
| Search | OpenSearch |
| Message Broker | Amazon MQ (managed RabbitMQ) |
| Image Storage | S3 |
| Email | SES |
| Push Notifications | SNS |
| Image Moderation | Rekognition |

---

## 🗄️ Database Per Service

| Service | Write DB | Read / Cache |
|---|---|---|
| Identity | PostgreSQL | Redis (sessions) |
| Catalog | PostgreSQL | OpenSearch |
| Commerce | PostgreSQL | Redis (cart) |
| Search | — | OpenSearch |
| Notification | PostgreSQL | — |
| Fulfillment | PostgreSQL | — |
| Moderation | PostgreSQL | — |

---

## 🔑 Key Architectural Patterns

### Saga Pattern (Distributed Transactions)
Used for the checkout flow:
```
Order Created → Reserve Listing → Process Payment → Confirm Order
                                         ↓ (on failure)
                              Compensating Transactions
```

### Outbox Pattern
Guarantees at-least-once event delivery:
1. Write domain events to an `outbox` table **in the same DB transaction**
2. A relay process reads the outbox and publishes to RabbitMQ

### Idempotent Consumers
All RabbitMQ consumers deduplicate by event ID using Redis to safely handle message redelivery.

### Tyk API Gateway Policies
```
Guest:  30 req/min
User:   200 req/min
Seller: 500 req/min
```
JWT middleware validates tokens issued by the Identity service.

---

## 🖥️ Frontend Architecture

```
src/
├── routes/                    # React Router v7 file-based routes
│   ├── _layout.tsx
│   ├── index.tsx              # Homepage / featured listings
│   ├── listings/
│   │   ├── index.tsx          # Search & browse
│   │   ├── $listingId.tsx     # Listing detail
│   │   └── new.tsx            # Create listing
│   ├── orders/
│   ├── profile/
│   └── admin/
├── features/                  # Feature-sliced architecture
│   ├── listings/
│   │   ├── api/
│   │   ├── components/
│   │   └── hooks/
│   ├── cart/
│   ├── auth/
│   └── search/
├── shared/
│   ├── components/            # shadcn/ui wrappers
│   ├── hooks/
│   └── lib/                   # API client, utils
└── store/                     # Zustand / React Context
```

**Key Frontend Patterns:**
- React Router v7 loaders for server-side data fetching
- Optimistic UI updates for cart operations
- WebSocket for real-time in-app notifications
- Image lazy loading with CloudFront CDN URLs

---

## 🚀 CI/CD Pipeline

### Environments
| Env | Trigger | Resources |
|---|---|---|
| `dev` | Push to `main` | 0.25 vCPU ECS tasks, Single-AZ RDS |
| `prod` | Release tag (`v*`) + Manual approval gate | Multi-AZ RDS, ECS autoscaling, CloudWatch alarms |

### Pipeline Stages
```
test  →  build  →  push to ECR  →  deploy-dev  →  [manual gate]  →  deploy-prod
```

- **Deploy prod:** Blue/Green deployment via ECS
- **Safety:** Automated smoke tests post-deploy, auto-rollback on health check failure

---

## 📅 Build Roadmap

### Phase 1 — MVP
1. Identity Service (auth foundation)
2. Catalog Service (create/view listings)
3. Search Service (browse & filter)
4. Basic React frontend

### Phase 2 — Commerce
5. Commerce Service (cart + orders + Stripe)
6. Notification Service (email confirmations)
7. Fulfillment Service (basic tracking)

### Phase 3 — Trust & Scale
8. Moderation Service (reports + AI image check)
9. Analytics / Admin dashboard
10. Rating & Review system

---

## 🗂️ Golang Service Folder Structure (per service)

```

order-service/
├── cmd/
│   ├── app/
│   │   └── main.go
│   └── migration/
│       └── main.go
│
├── config/
│   └── config.go
│
├── internal/
│   ├── bootstrap/
│   │   └── app.go
│   │
│   ├── domain/                                  # 🔴 DDD: Core — zero external dependencies
│   │   ├── aggregate/
│   │   │   ├── order.go
│   │   │   └── order_item.go
│   │   ├── entity/
│   │   │   └── buyer.go
│   │   ├── valueobject/
│   │   │   ├── money.go
│   │   │   ├── address.go
│   │   │   └── order_status.go
│   │   ├── event/                               # 🔵 EDA: Domain events definitions
│   │   │   ├── order_created.go
│   │   │   ├── order_cancelled.go
│   │   │   └── order_shipped.go
│   │   ├── repository/
│   │   │   └── order_repository.go
│   │   └── service/
│   │       └── pricing_service.go
│   │
│   ├── application/                             # 🟡 Clean Arch: Use-case layer
│   │   ├── command/                             # 🟣 CQRS: Write side
│   │   │   ├── create_order/
│   │   │   │   ├── command.go
│   │   │   │   └── handler.go
│   │   │   ├── cancel_order/
│   │   │   │   ├── command.go
│   │   │   │   └── handler.go
│   │   │   └── ship_order/
│   │   │       ├── command.go
│   │   │       └── handler.go
│   │   ├── query/                               # 🟣 CQRS: Read side
│   │   │   ├── get_order/
│   │   │   │   ├── query.go
│   │   │   │   └── handler.go
│   │   │   └── list_orders/
│   │   │       ├── query.go
│   │   │       └── handler.go
│   │   ├── eventhandler/                        # 🔵 EDA: Handles incoming RabbitMQ events
│   │   │   ├── handler.go
│   │   │   ├── payment_confirmed.go
│   │   │   └── inventory_reserved.go
│   │   ├── port/
│   │   │   ├── event_publisher.go
│   │   │   └── read_model_store.go
│   │   └── dto/
│   │       ├── order_request.go
│   │       └── order_response.go
│   │
│   ├── infrastructure/                          # 🟢 Clean Arch: Adapters / outer layer
│   │   ├── persistence/
│   │   │   └── postgres/
│   │   │       ├── postgres.go
│   │   │       ├── order_repository.go
│   │   │       └── migrations/
│   │   │           ├── 001_create_orders.sql
│   │   │           └── 002_add_buyer.sql
│   │   ├── messaging/
│   │   │   └── rabbitmq/
│   │   │       ├── connection.go
│   │   │       └── consumer.go
│   │   └── server/
│   │       └── http.go
│   │
│   ├── interface/                               # 🟠 Clean Arch: Delivery layer
│   │   └── http/
│   │       ├── router.go
│   │       ├── handler/
│   │       │   └── order_handler.go
│   │       └── middleware/
│   │           ├── chain.go
│   │           ├── logger.go
│   │           └── request_id.go
│   │
│   └── errors/                                  # ✅ NEW: Sentinel & domain errors
│       ├── errors.go                            # Base error types & wrapping helpers
│       ├── domain_error.go                      # Business rule violations (e.g. invalid state)
│       ├── not_found_error.go                   # Resource not found
│       └── internal_error.go                    # Unexpected system/infra errors
│
├── test/                                        # ✅ NEW: BDD tests (outside internal — tests all layers)
│   ├── suite_test.go                            # Global suite setup & teardown
│   ├── features/                                # Gherkin .feature files (the "what")
│   │   ├── create_order.feature
│   │   ├── cancel_order.feature
│   │   └── get_order.feature
│   ├── steps/                                   # Step definitions (the "how")
│   │   ├── order_steps.go
│   │   └── common_steps.go
│   └── helper/                                  # Test utilities & fixtures
│       ├── fixtures.go                          # Seed data builders
│       ├── db_helper.go                         # Test DB setup/teardown
│       └── mq_helper.go                         # RabbitMQ test helpers
│
├── env.json
├── env.example.json
├── env.test.json
├── .gitignore
└── go.mod
```


---

## 💡 Conversation Notes

- **Location context:** Ho Chi Minh City, Vietnam → shipping providers include GHN and GHTK
- **Project type:** Personal project (start lean, grow into full architecture)
- **Next steps to discuss:** Any specific service implementation details, data models, API contracts, or frontend component design

---

*Generated: February 26, 2026 | Stack: Go + React v19 + Tyk + RabbitMQ + AWS*
