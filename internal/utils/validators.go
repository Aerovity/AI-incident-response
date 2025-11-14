package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email is valid
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateRequired checks if a string is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateStringLength checks if a string is within length bounds
func ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(value)
	if length < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, max)
	}
	return nil
}

// ValidateSlug checks if a slug is valid (lowercase letters, numbers, hyphens only)
func ValidateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug is required")
	}
	matched := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(slug)
	if !matched {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}
