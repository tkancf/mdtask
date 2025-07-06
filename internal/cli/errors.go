package cli

import (
	"fmt"
	"os"

	"github.com/tkancf/mdtask/internal/errors"
)

// HandleError handles errors consistently across CLI commands
func HandleError(err error, format string) {
	if err == nil {
		return
	}

	// For JSON output, print error in JSON format
	if format == "json" {
		fmt.Fprintf(os.Stderr, `{"error": "%s"}`+"\n", err.Error())
		if errors.IsNotFound(err) {
			os.Exit(1)
		} else if errors.IsInvalidInput(err) {
			os.Exit(2)
		} else if errors.IsPermission(err) {
			os.Exit(3)
		} else if errors.IsDuplicate(err) {
			os.Exit(4)
		} else {
			os.Exit(5)
		}
		return
	}

	// For text output, print to stderr
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	
	// Exit with appropriate code
	if errors.IsNotFound(err) {
		os.Exit(1)
	} else if errors.IsInvalidInput(err) {
		os.Exit(2)
	} else if errors.IsPermission(err) {
		os.Exit(3)
	} else if errors.IsDuplicate(err) {
		os.Exit(4)
	} else {
		os.Exit(5)
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}