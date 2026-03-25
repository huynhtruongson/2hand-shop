package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	appErr "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/http/types"
)

func ResponseError(c *gin.Context, err error) {
	c.Errors = append(c.Errors, c.Error(err))
	if ae, ok := appErr.As(err); ok {
		userFacing := ae.UserFacing()
		c.JSON(ae.HTTPStatus(), types.HttpResponse{
			Success: false,
			Message: userFacing.Code,
			Error: &types.ErrorResponse{
				Code:    userFacing.Code,
				Message: userFacing.Message,
				Details: userFacing.Details,
			},
		})
		return
	}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string, len(ve))
		for _, fe := range ve {
			details[strings.ToLower(fe.Field())] = validationMessage(fe)
		}
		c.JSON(http.StatusBadRequest, types.HttpResponse{
			Success: false,
			Message: "Validation failed",
			Error: &types.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: "One or more fields failed validation",
				Details: details,
			},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, types.HttpResponse{
		Success: false,
		Message: "Internal server error",
		Error: &types.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
		}})
}

func Response(c *gin.Context, data any, message ...string) {
	msg := "ok"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusOK, types.HttpResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func validationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + fe.Param()
	case "max":
		return field + " must be at most " + fe.Param()
	case "len":
		return field + " must be exactly " + fe.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	default:
		return field + " is invalid"
	}
}
