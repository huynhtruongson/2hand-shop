package errors

import "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"

// ── Generic / cross-cutting ──────────────────────────────────────────────────

var (
	ErrInternal = errors.NewAppError(errors.KindInternal, "INTERNAL_ERROR", "an unexpected error occurred")

	ErrUnavailable = errors.NewAppError(errors.KindUnavailable, "SERVICE_UNAVAILABLE", "service is temporarily unavailable, please try again later")

	ErrRateLimit = errors.NewAppError(errors.KindRateLimit, "RATE_LIMIT_EXCEEDED", "too many requests, please slow down")

	ErrValidation = errors.NewAppError(errors.KindValidation, "VALIDATION_ERROR", "request validation failed")

	ErrBadRequest = errors.NewAppError(errors.KindBadRequest, "BAD_REQUEST", "invalid request format")

	ErrUnprocessable = errors.NewAppError(errors.KindUnprocessable, "UNPROCESSABLE", "request could not be processed")

	ErrRecordNotFound = errors.NewAppError(errors.KindNotFound, "RECORD_NOT_FOUND", "record not found")
)

// ── Catalog / Product ────────────────────────────────────────────────────────

var (
	ErrProductNotFound = errors.NewAppError(errors.KindNotFound, "PRODUCT_NOT_FOUND", "product not found")

	ErrProductAlreadyExists = errors.NewAppError(errors.KindConflict, "PRODUCT_ALREADY_EXISTS", "a product with this id already exists")

	ErrProductInvalidStatusTransition = errors.NewAppError(errors.KindBadRequest, "PRODUCT_INVALID_STATUS_TRANSITION", "invalid product status transition")

	ErrProductNotActive = errors.NewAppError(errors.KindBadRequest, "PRODUCT_NOT_PUBLISHED", "product is not in published status")
)

// ── Catalog / ProductRequest ─────────────────────────────────────────────────

var (
	ErrProductRequestNotFound = errors.NewAppError(errors.KindNotFound, "PRODUCT_REQUEST_NOT_FOUND", "product request not found")

	ErrProductRequestAlreadyExists = errors.NewAppError(errors.KindConflict, "PRODUCT_REQUEST_ALREADY_EXISTS", "a product request with this id already exists")

	ErrProductRequestNotEditable = errors.NewAppError(errors.KindForbidden, "PRODUCT_REQUEST_NOT_EDITABLE", "product request cannot be edited because it is no longer in pending status")
)

// ── Catalog / Category ───────────────────────────────────────────────────────

var (
	ErrCategoryNotFound = errors.NewAppError(errors.KindNotFound, "CATEGORY_NOT_FOUND", "category not found")

	ErrCategoryAlreadyExists = errors.NewAppError(errors.KindConflict, "CATEGORY_ALREADY_EXISTS", "a category with this id already exists")
)
