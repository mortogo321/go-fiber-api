package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidateStruct validates a struct using go-playground/validator tags.
// Returns nil if the struct is valid, or a slice of ValidationError otherwise.
func ValidateStruct(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   e.Field(),
			Tag:     e.Tag(),
			Message: formatMessage(e),
		})
	}
	return errors
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return e.Field() + " must be a valid email"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	default:
		return e.Field() + " failed on " + e.Tag() + " validation"
	}
}
