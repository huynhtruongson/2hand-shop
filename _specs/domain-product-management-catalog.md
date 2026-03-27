# Spec for Domain Product Management (Catalog Service)

branch: feature/domain-product-management-catalog
figma_component (if used):

## Summary

This spec defines the product domain model in the Catalog Service for a second-hand marketplace. It introduces two primary aggregates — `Product` (admin-owned inventory) and `SellRequest` (client-initiated selling intent) — alongside an admin approval workflow that governs how client-submitted items become shop products.

## Functional Requirements

### Product Aggregate

- A `Product` is an admin-owned item available for sale in the marketplace.
- A `Product` must have the following fields:
  - `ID` (UUID) — unique identifier
  - `Name` (string, non-empty) — product title/name
  - `Description` (string) — free-text description
  - `Category` (enum: `clothes`, `accessories`, `shoes`) — only these three values are allowed
  - `Image` (URL or attachment) — primary product image; additional images supported
  - `Size` (string, free text) — e.g. "M", "42", "One Size"
  - `Gender` (enum: `unisex`, `male`, `female`) — intended gender audience
  - `Condition` (enum: `new`, `like_new`, `good`, `fair`, `poor`) — quality grade
  - `Price` (decimal, positive) — selling price in the shop's currency
  - `Brand` (string) — brand name; free text (empty allowed)
  - `OwnerID` (UUID) — references the admin user who owns/controls this product
  - `Status` (enum: `active`, `sold`, `archived`) — lifecycle state
  - `CreatedAt`, `UpdatedAt` (timestamps)
- A `Product` can only be mutated (updated, archived) by its owning admin.
- Status transitions for `Product`: `active` → `sold` | `archived`. No backward transitions.

### SellRequest Aggregate

- A `SellRequest` is a client-initiated intent to sell an item to the shop.
- A `SellRequest` must have the following fields:
  - `ID` (UUID) — unique identifier
  - `ProductDetails` (embedded/nested object) — mirrors the product fields: `Name`, `Description`, `Category`, `Image`, `Size`, `Gender`, `Condition`, `ExpectedPrice`, `Brand`
  - `ContactInfo` — client's contact details: `Name`, `Email`, `Phone` (at least one required)
  - `RequesterID` (UUID) — the client user who submitted the request
  - `Status` (enum: `pending`, `approved`, `rejected`) — lifecycle state
  - `RejectionReason` (string, optional) — reason provided by admin if rejected
  - `CreatedAt`, `UpdatedAt` (timestamps)
- `ExpectedPrice` is the client's desired selling price; the admin sets the final shop `Price` upon approval.

### SellRequest Lifecycle

- A `SellRequest` starts in `pending` status.
- **Approved path**: Admin reviews and approves the request. A new `Product` is created from the request's `ProductDetails`. The `Product.Price` is set by the admin (not necessarily equal to `ExpectedPrice`). The admin becomes the `OwnerID` of the new `Product`. The `SellRequest` transitions to `approved`.
- **Rejected path**: Admin rejects the request with an optional reason. The `SellRequest` transitions to `rejected`. A rejected `SellRequest` is a **dead end** — it cannot be reopened or edited. The client must submit a new `SellRequest` if they wish to try again.
- Once in `approved` or `rejected`, a `SellRequest` is immutable.

### Admin Operations

- Admin can create a `Product` directly (without a `SellRequest`) — for shop-owned inventory.
- Admin can update an existing `Product` (name, description, price, status, etc.).
- Admin can archive or mark a `Product` as `sold`.
- Admin can list all `SellRequest`s (with filter by status).
- Admin can approve a `pending` `SellRequest` (supplying the final price and additional product details).
- Admin can reject a `pending` `SellRequest` (supplying an optional reason).

### Client Operations

- Authenticated client can submit a new `SellRequest`.
- Authenticated client can view their own `SellRequest`s (with status).
- Authenticated client can cancel their own `pending` `SellRequest` before it is reviewed.
- Authenticated client **cannot** approve, reject, or directly create `Product`s.

## Figma Design Reference (only if referenced)

## Possible Edge Cases

- Client submits a `SellRequest` for a `Category` not in the allowed set — validation rejects the request.
- Admin approves a `SellRequest` but sets `Price` to zero or negative — validation prevents this.
- Client attempts to cancel a `SellRequest` that is already `approved` or `rejected` — operation is denied.
- Admin approves a `SellRequest` but the same `RequesterID` already has an `active` `SellRequest` for a nearly identical item — no automatic duplicate detection; admin must judge manually.
- `SellRequest` with incomplete `ContactInfo` — at least one contact field (`Email` or `Phone`) is required.

## Acceptance Criteria

- A `Product` can only be created via admin action (direct creation or approval of a `SellRequest`).
- A `SellRequest` transitions from `pending` → `approved` (generates a `Product`) or `pending` → `rejected` (dead end).
- Rejected `SellRequest`s cannot be modified or resubmitted; client must create a new one.
- Only admins can approve or reject `SellRequest`s.
- Only the owning admin can mutate their `Product`s.
- `Product.Status` follows: `active` → `sold` | `archived` with no backward transitions.
- All domain rules are enforced at the aggregate level; repositories enforce persistence constraints.
- Domain events are published for: `SellRequest` submitted, approved, rejected; `Product` created, updated, sold, archived.

## Open Questions

- Should `Image` support multiple images per product/sell request, or a single primary image?
- Is there a maximum number of images per product?
- Does `Brand` need to be validated against a predefined list, or is free text acceptable?
- Should there be a soft delete (archive) vs. hard delete for `Product`s?
- What notification (email/push/in-app) should be sent to the client when their `SellRequest` is approved or rejected?
- Does the admin need a bulk approve/reject capability for multiple `SellRequest`s at once?
- Should `SellRequest` include an optional field for the client to upload multiple images?

## Testing Guidelines

- Unit tests for `Product` aggregate: valid creation, invalid field combinations, status transition enforcement, ownership validation.
- Unit tests for `SellRequest` aggregate: valid submission, contact info validation, status transition rules (pending → approved, pending → rejected, immutability after terminal states).
- Unit tests for domain services covering admin approve/reject logic.
- Integration tests for the full approve flow: `SellRequest` approved → `Product` created with correct fields and ownership.
- BDD scenarios covering: happy path for client submit, admin approve; client cancel pending request; admin reject with reason; client attempts to re-submit rejected request.
