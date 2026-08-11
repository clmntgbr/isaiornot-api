package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func Validator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			if name := tagName(fld, "json"); name != "" {
				return name
			}
			if name := tagName(fld, "query"); name != "" {
				return name
			}
			return fld.Name
		})
	})
	return validate
}

func tagName(fld reflect.StructField, key string) string {
	name := strings.SplitN(fld.Tag.Get(key), ",", 2)[0]
	if name == "" || name == "-" {
		return ""
	}
	return name
}

func Struct(value any) error {
	return Validator().Struct(value)
}

// FieldErrors maps json/query field names to validation messages.
func FieldErrors(err error) map[string]string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string]string{"_error": err.Error()}
	}

	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		field := fe.Field()
		if ns := fe.Namespace(); ns != "" {
			if i := strings.Index(ns, "."); i >= 0 {
				field = ns[i+1:]
			}
		}
		out[field] = messageFor(fe)
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "uuid":
		return "must be a valid UUID"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "dive":
		return "contains invalid items"
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}

func BindAndValidateBody(c fiber.Ctx, dst any) error {
	if err := c.Bind().Body(dst); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}
	if err := Struct(dst); err != nil {
		return ValidationFailed(c, err)
	}
	return nil
}

func BindAndValidateQuery(c fiber.Ctx, dst any) error {
	if err := c.Bind().Query(dst); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	if err := Struct(dst); err != nil {
		return ValidationFailed(c, err)
	}
	return nil
}

func ValidationFailed(c fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "Validation failed",
		"errors":  FieldErrors(err),
	})
}
