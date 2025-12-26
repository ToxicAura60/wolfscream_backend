package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func init() {
	Validate.RegisterValidation("snakecase", snakeCase)
}

func snakeCase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9_]*$`, value)
	return matched
}
