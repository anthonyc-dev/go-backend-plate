package utils

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("valid-email", validateEmail)
	validate.RegisterValidation("valid-name", validateName)
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return len(name) >= 2 && len(name) <= 100
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidationError formats validator errors into a JSON-friendly map
// It can also wrap custom errors (like "user not found")
func ValidationError(err error) gin.H {
	errors := make(map[string]string)

	if err == nil {
		return gin.H{}
	}

	// If it's a validator error
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			// e.Field() = struct field, e.Tag() = validation rule
			errors[e.Field()] = e.Tag()
		}
		return gin.H{"errors": errors}
	}

	// For custom or general errors
	errors["error"] = err.Error()
	return gin.H{"errors": errors}
}
