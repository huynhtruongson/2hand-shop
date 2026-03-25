package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		assertFn func(t *testing.T, err error)
	}{
		{
			name: "bad request error",
			err:  NewAppError(KindBadRequest, "INVALID_PRODUCT_REQUEST", "invalid product request"),
			assertFn: func(t *testing.T, err error) {
				ae, ok := As(err)
				assert.True(t, ok)
				assert.Equal(t, ae.Kind(), KindBadRequest)
				assert.Equal(t, ae.HTTPStatus(), http.StatusBadRequest)
				assert.Equal(t, ae.Code(), "INVALID_PRODUCT_REQUEST")
				assert.Equal(t, ae.Error(), "[INVALID_PRODUCT_REQUEST] invalid product request")
				userError := ae.UserFacing()
				assert.Equal(t, userError.Code, "INVALID_PRODUCT_REQUEST")
				assert.Equal(t, userError.Message, "invalid product request")
			},
		},
		{
			name: "unauthorized error",
			err:  NewAppError(KindUnauthorized, "UNAUTHORIZED_ACCESS", "unauthorized access"),
			assertFn: func(t *testing.T, err error) {
				ae, ok := As(err)
				assert.True(t, ok)
				assert.Equal(t, ae.Kind(), KindUnauthorized)
				assert.Equal(t, ae.HTTPStatus(), http.StatusUnauthorized)
				assert.Equal(t, ae.Code(), "UNAUTHORIZED_ACCESS")
				assert.Equal(t, ae.Error(), "[UNAUTHORIZED_ACCESS] unauthorized access")
				userError := ae.UserFacing()
				assert.Equal(t, userError.Code, "UNAUTHORIZED_ACCESS")
				assert.Equal(t, userError.Message, "unauthorized access")
			},
		},
		{
			name: "not found error",
			err: NewAppError(KindNotFound, "PRODUCT_NOT_FOUND", "product not found").WithDetails(
				map[string]string{
					"product_id": "product-123",
				},
			),
			assertFn: func(t *testing.T, err error) {
				ae, ok := As(err)
				assert.True(t, ok)
				assert.Equal(t, ae.Kind(), KindNotFound)
				assert.Equal(t, ae.HTTPStatus(), http.StatusNotFound)
				assert.Equal(t, ae.Code(), "PRODUCT_NOT_FOUND")
				assert.Equal(t, ae.Error(), "[PRODUCT_NOT_FOUND] product not found")
				userError := ae.UserFacing()
				assert.Equal(t, userError.Code, "PRODUCT_NOT_FOUND")
				assert.Equal(t, userError.Message, "product not found")
				assert.Equal(t, userError.Details, map[string]string{
					"product_id": "product-123",
				})
			},
		},
		{
			name: "internal error",
			err: NewAppError(KindInternal, "INTERNAL_ERROR", "internal error").
				WithInternal("db connection error").WithCause(errors.New("db connection closed")),
			assertFn: func(t *testing.T, err error) {
				ae, ok := As(err)
				assert.True(t, ok)
				assert.Equal(t, ae.Kind(), KindInternal)
				assert.Equal(t, ae.HTTPStatus(), http.StatusInternalServerError)
				assert.Equal(t, ae.Code(), "INTERNAL_ERROR")
				assert.Equal(t, ae.Error(), "[INTERNAL_ERROR] db connection error: db connection closed")
				userError := ae.UserFacing()
				assert.Equal(t, userError.Code, "INTERNAL_ERROR")
				assert.Equal(t, userError.Message, "internal error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.assertFn(t, tt.err)
		})
	}
}
