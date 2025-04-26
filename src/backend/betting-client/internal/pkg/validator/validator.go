package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		var errorMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMsgs = append(errorMsgs, fmt.Sprintf("Field '%s' failed validation on '%s'", err.Field(), err.Tag()))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errorMsgs, "; "))
	}
	return nil
}

func GetValidator() *validator.Validate {
	return validate
}
