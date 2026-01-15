package validator

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// ValidateUUID validates if a string is a valid UUID
func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ValidateNotEmpty validates if a string is not empty (after trimming)
func ValidateNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// ValidateLength validates if a string length is within the specified range
func ValidateLength(s string, min, max int) bool {
	length := len(strings.TrimSpace(s))
	return length >= min && length <= max
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	return ValidateNotEmpty(parts[0]) && ValidateNotEmpty(parts[1])
}

// SanitizeString removes leading/trailing whitespace and normalizes
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// ValidateMessageContent validates message content
func ValidateMessageContent(content string, maxLength int) error {
	content = SanitizeString(content)
	if content == "" {
		return &ValidationError{Field: "content", Message: "message content cannot be empty"}
	}
	if len(content) > maxLength {
		return &ValidationError{Field: "content", Message: "message content exceeds maximum length"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// HasOnlyPrintable checks if string contains only printable characters
func HasOnlyPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

