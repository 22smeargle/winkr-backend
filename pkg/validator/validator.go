package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init initializes the validator
func Init() {
	validate = validator.New()
	
	// Register custom validation tags
	validate.RegisterValidation("password", validatePassword)
}

// Validate validates a struct and returns validation errors
func Validate(s interface{}) error {
	if validate == nil {
		Init()
	}
	
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	
	// Convert validation errors to a more readable format
	validationErrors := err.(validator.ValidationErrors)
	var errorMessages []string
	
	for _, e := range validationErrors {
		errorMessages = append(errorMessages, getErrorMessage(e))
	}
	
	return fmt.Errorf(strings.Join(errorMessages, "; "))
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	if validate == nil {
		Init()
	}
	
	err := validate.Var(field, tag)
	if err == nil {
		return nil
	}
	
	validationErrors := err.(validator.ValidationErrors)
	var errorMessages []string
	
	for _, e := range validationErrors {
		errorMessages = append(errorMessages, getErrorMessage(e))
	}
	
	return fmt.Errorf(strings.Join(errorMessages, "; "))
}

// getErrorMessage converts a validation error to a human-readable message
func getErrorMessage(e validator.FieldError) string {
	fieldName := e.Field()
	
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fieldName, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fieldName, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fieldName, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fieldName, e.Param())
	case "password":
		return fmt.Sprintf("%s must contain at least 8 characters, including uppercase, lowercase, number, and special character", fieldName)
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

// validatePassword is a custom validator for password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// GetValidator returns the underlying validator instance
func GetValidator() *validator.Validate {
	if validate == nil {
		Init()
	}
	return validate
}

// RegisterValidation registers a custom validation function
func RegisterValidation(tag string, fn validator.Func) error {
	if validate == nil {
		Init()
	}
	return validate.RegisterValidation(tag, fn)
}

// RegisterValidationCtx registers a custom validation function with context
func RegisterValidationCtx(tag string, fn validator.FuncCtx) error {
	if validate == nil {
		Init()
	}
	return validate.RegisterValidation(tag, fn)
}

// RegisterAlias registers a validation alias
func RegisterAlias(alias, tags string) error {
	if validate == nil {
		Init()
	}
	return validate.RegisterAlias(alias, tags)
}

// RegisterTagNameFunc registers a function to get custom tag names
func RegisterTagNameFunc(fn validator.StructTagNameFunc) {
	if validate == nil {
		Init()
	}
	validate.RegisterTagNameFunc(fn)
}

// StructExcept validates a struct except for the specified fields
func StructExcept(s interface{}, fields ...string) error {
	if validate == nil {
		Init()
	}
	
	// Create a map of fields to exclude
	exclude := make(map[string]bool)
	for _, field := range fields {
		exclude[field] = true
	}
	
	// Get the struct value
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Iterate through struct fields
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		
		// Skip excluded fields
		if exclude[fieldName] {
			continue
		}
		
		// Validate the field
		fieldValue := val.Field(i)
		if err := validate.Var(fieldValue.Interface(), field.Tag.Get("validate")); err != nil {
			return err
		}
	}
	
	return nil
}

// StructOnly validates only the specified fields in a struct
func StructOnly(s interface{}, fields ...string) error {
	if validate == nil {
		Init()
	}
	
	// Get the struct value
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Create a map of fields to include
	include := make(map[string]bool)
	for _, field := range fields {
		include[field] = true
	}
	
	// Iterate through struct fields
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		
		// Skip non-included fields
		if !include[fieldName] {
			continue
		}
		
		// Validate the field
		fieldValue := val.Field(i)
		if err := validate.Var(fieldValue.Interface(), field.Tag.Get("validate")); err != nil {
			return err
		}
	}
	
	return nil
}