package libkv

import (
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"
)

const defaultSpace = "default"

func spacePath(space string) string {
	return space + "/"
}

// spaceValidator validates if field contains allowed characters for space name
func spaceValidator(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`).MatchString(fl.Field().String())
}
