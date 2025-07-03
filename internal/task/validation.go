package task

import (
	"fmt"
	"strings"
)

// ValidateTitle checks if the title is valid (no newlines)
func ValidateTitle(title string) error {
	if strings.Contains(title, "\n") || strings.Contains(title, "\r") {
		return fmt.Errorf("title cannot contain newlines")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	return nil
}

// ValidateDescription checks if the description is valid (no newlines)
func ValidateDescription(description string) error {
	if strings.Contains(description, "\n") || strings.Contains(description, "\r") {
		return fmt.Errorf("description cannot contain newlines")
	}
	return nil
}

// Validate validates the entire task
func (t *Task) Validate() error {
	if err := ValidateTitle(t.Title); err != nil {
		return err
	}
	if err := ValidateDescription(t.Description); err != nil {
		return err
	}
	return nil
}