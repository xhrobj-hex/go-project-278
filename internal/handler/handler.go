package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func handleBindError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": validationErrorsResponse(validationErrors),
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "invalid request",
	})
}

func validationErrorsResponse(validationErrors validator.ValidationErrors) gin.H {
	result := gin.H{}

	for _, fieldError := range validationErrors {
		field := validationFieldName(fieldError)
		result[field] = validationErrorMessage(fieldError)
	}

	return result
}

func validationFieldName(fieldError validator.FieldError) string {
	switch fieldError.Field() {
	case "OriginalURL":
		return "original_url"
	case "ShortName":
		return "short_name"
	default:
		return fieldError.Field()
	}
}

func validationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "url":
		return "must be a valid URL"
	case "min":
		return "must be at least " + fieldError.Param() + " characters"
	case "max":
		return "must be at most " + fieldError.Param() + " characters"
	default:
		return fieldError.Error()
	}
}

func validationErrorResponse(field, message string) gin.H {
	return gin.H{
		"errors": gin.H{
			field: message,
		},
	}
}
