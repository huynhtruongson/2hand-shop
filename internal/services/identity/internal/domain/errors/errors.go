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

// ── Auth / Identity ──────────────────────────────────────────────────────────

var (
	ErrUnauthorized = errors.NewAppError(errors.KindUnauthorized, "AUTH_UNAUTHORIZED", "authentication required")

	ErrTokenExpired = errors.NewAppError(errors.KindUnauthorized, "AUTH_TOKEN_EXPIRED", "your session has expired, please log in again")

	ErrTokenInvalid = errors.NewAppError(errors.KindUnauthorized, "AUTH_TOKEN_INVALID", "invalid authentication token")

	ErrForbidden = errors.NewAppError(errors.KindForbidden, "AUTH_FORBIDDEN", "you do not have permission to perform this action")

	ErrInvalidCredentials = errors.NewAppError(errors.KindUnauthorized, "AUTH_INVALID_CREDENTIALS", "invalid email or password")

	ErrExpiredConfirmationCode = errors.NewAppError(errors.KindUnprocessable, "AUTH_EXPIRED_CONFIRMATION_CODE", "confirmation code has expired")

	ErrInvalidConfirmationCode = errors.NewAppError(errors.KindUnprocessable, "AUTH_INVALID_CONFIRMATION_CODE", "invalid confirmation code")

	ErrUserNotVerified = errors.NewAppError(errors.KindUnprocessable, "USER_NOT_VERIFIED", "user not verified")

	ErrUserNotFound = errors.NewAppError(errors.KindNotFound, "USER_NOT_FOUND", "user not found")

	ErrUserAlreadyExists = errors.NewAppError(errors.KindConflict, "USER_ALREADY_EXISTS", "a user with this email already exists")

	ErrUserSuspended = errors.NewAppError(errors.KindForbidden, "USER_SUSPENDED", "your account has been suspended")
)
