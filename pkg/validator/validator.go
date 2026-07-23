package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	v *validator.Validate
}

func New() *CustomValidator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &CustomValidator{v: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}
