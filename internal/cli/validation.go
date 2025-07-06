package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/tkancf/mdtask/internal/constants"
	"github.com/tkancf/mdtask/internal/task"
)

// NormalizeTaskID ensures task ID has the proper prefix
func NormalizeTaskID(id string) (string, error) {
	if strings.HasPrefix(id, constants.TaskIDPrefix) {
		return id, nil
	}
	
	// Try to parse as timestamp
	if _, err := time.Parse("20060102150405", id); err == nil {
		return constants.TaskIDPrefix + id, nil
	}
	
	// Check if it includes suffix (e.g., "20240101120000_1")
	parts := strings.Split(id, "_")
	if len(parts) >= 1 {
		if _, err := time.Parse("20060102150405", parts[0]); err == nil {
			return constants.TaskIDPrefix + id, nil
		}
	}
	
	return "", fmt.Errorf("invalid task ID format: %s", id)
}

// ValidateStatus checks if the given status string is valid
func ValidateStatus(status string) (task.Status, error) {
	switch task.Status(status) {
	case task.StatusTODO, task.StatusWIP, task.StatusWAIT, task.StatusSCHE, task.StatusDONE:
		return task.Status(status), nil
	default:
		return "", fmt.Errorf("invalid status: %s (valid: TODO, WIP, WAIT, SCHE, DONE)", status)
	}
}

// ParseDeadline parses deadline string in YYYY-MM-DD format
func ParseDeadline(deadline string) (*time.Time, error) {
	if deadline == "" {
		return nil, nil
	}
	
	t, err := time.Parse("2006-01-02", deadline)
	if err != nil {
		return nil, fmt.Errorf("invalid deadline format (use YYYY-MM-DD): %s", deadline)
	}
	
	return &t, nil
}

// ParseReminder parses reminder string in YYYY-MM-DD HH:MM format
func ParseReminder(reminder string) (*time.Time, error) {
	if reminder == "" {
		return nil, nil
	}
	
	t, err := time.Parse("2006-01-02 15:04", reminder)
	if err != nil {
		return nil, fmt.Errorf("invalid reminder format (use YYYY-MM-DD HH:MM): %s", reminder)
	}
	
	return &t, nil
}