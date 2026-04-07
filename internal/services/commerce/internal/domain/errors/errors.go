package errors

import "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"

// ── Auth  ──────────────────────────────────────────────────────────────────

var (
	ErrUnauthorized = errors.NewAppError(errors.KindUnauthorized, "AUTH_UNAUTHORIZED", "authentication required")
)

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

// ── Commerce / Order ─────────────────────────────────────────────────────────

var (
	ErrOrderNotFound = errors.NewAppError(errors.KindNotFound, "ORDER_NOT_FOUND", "order not found")

	ErrOrderAlreadyExists = errors.NewAppError(errors.KindConflict, "ORDER_ALREADY_EXISTS", "an order with this id already exists")

	ErrOrderInvalidStatusTransition = errors.NewAppError(errors.KindBadRequest, "ORDER_INVALID_STATUS_TRANSITION", "invalid order status transition")

	ErrOrderNotEditable = errors.NewAppError(errors.KindForbidden, "ORDER_NOT_EDITABLE", "order cannot be edited because it is no longer in a mutable state")
)

// ── Commerce / Cart ─────────────────────────────────────────────────────────

var (
	ErrCartNotFound = errors.NewAppError(errors.KindNotFound, "CART_NOT_FOUND", "cart not found")

	ErrCartEmpty = errors.NewAppError(errors.KindBadRequest, "CART_EMPTY", "cannot create order from empty cart")

	ErrCartItemNotFound = errors.NewAppError(errors.KindNotFound, "CART_ITEM_NOT_FOUND", "cart item not found")
)

// ── Commerce / Payment ──────────────────────────────────────────────────────

var (
	ErrPaymentNotFound = errors.NewAppError(errors.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")

	ErrPaymentAlreadyExists = errors.NewAppError(errors.KindConflict, "PAYMENT_ALREADY_EXISTS", "a payment with this id already exists")

	ErrPaymentAlreadyConfirmed = errors.NewAppError(errors.KindConflict, "PAYMENT_ALREADY_CONFIRMED", "payment has already been confirmed")

	ErrPaymentInvalidTransition = errors.NewAppError(errors.KindBadRequest, "PAYMENT_INVALID_STATUS_TRANSITION", "invalid payment status transition")

	ErrPaymentFailed = errors.NewAppError(errors.KindBadRequest, "PAYMENT_FAILED", "payment processing failed")
)