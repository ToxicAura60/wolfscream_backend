package validator

import "github.com/go-playground/validator/v10"

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatError(err error) []FieldError {
	errs := err.(validator.ValidationErrors)
	result := []FieldError{}

	for _, e := range errs {
		result = append(result, FieldError{
			Field:   e.Field(),
			Message: message(e),
		})
	}

	return result
}

func message(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "snakecase":
		return "must be snake_case lowercase without spaces"
	case "max":
		return "value too long"
	default:
		return "invalid value"
	}
}
