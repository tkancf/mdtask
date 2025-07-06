package task

import (
	"strings"
	"time"

	"github.com/tkancf/mdtask/internal/constants"
)

type Status string

const (
	StatusTODO Status = Status(constants.StatusTODO)
	StatusWIP  Status = Status(constants.StatusWIP)
	StatusWAIT Status = Status(constants.StatusWAIT)
	StatusSCHE Status = Status(constants.StatusSCHE)
	StatusDONE Status = Status(constants.StatusDONE)
)

type Task struct {
	ID          string
	Title       string
	Description string
	Aliases     []string
	Tags        []string
	Created     time.Time
	Updated     time.Time
	Content     string
}

// Helper function to filter tags by prefix
func (t *Task) filterTagsByPrefix(prefix string, exclude bool) []string {
	var result []string
	for _, tag := range t.Tags {
		hasPrefix := strings.HasPrefix(tag, prefix)
		if (hasPrefix && !exclude) || (!hasPrefix && exclude) {
			result = append(result, tag)
		}
	}
	return result
}

// Helper function to get tag value with prefix
func (t *Task) getTagWithPrefix(prefix string) (string, bool) {
	prefixLen := len(prefix)
	for _, tag := range t.Tags {
		if len(tag) > prefixLen && strings.HasPrefix(tag, prefix) {
			return tag[prefixLen:], true
		}
	}
	return "", false
}

// Helper function to set tag with prefix
func (t *Task) setTagWithPrefix(prefix, value string) {
	t.Tags = t.filterTagsByPrefix(prefix, true)
	if value != "" {
		t.Tags = append(t.Tags, prefix+value)
	}
}

// Helper function to check if tag exists
func (t *Task) hasTag(tag string) bool {
	for _, t := range t.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (t *Task) GetStatus() Status {
	if value, ok := t.getTagWithPrefix(constants.StatusTagPrefix); ok {
		return Status(value)
	}
	return StatusTODO
}

func (t *Task) SetStatus(status Status) {
	t.setTagWithPrefix(constants.StatusTagPrefix, string(status))
}

func (t *Task) IsArchived() bool {
	return t.hasTag(constants.ArchivedTag)
}

func (t *Task) Archive() {
	if !t.IsArchived() {
		t.Tags = append(t.Tags, constants.ArchivedTag)
	}
}

func (t *Task) Unarchive() {
	var newTags []string
	for _, tag := range t.Tags {
		if tag != constants.ArchivedTag {
			newTags = append(newTags, tag)
		}
	}
	t.Tags = newTags
}

func (t *Task) GetDeadline() *time.Time {
	if value, ok := t.getTagWithPrefix(constants.DeadlineTagPrefix); ok {
		if deadline, err := time.Parse(constants.DateFormat, value); err == nil {
			return &deadline
		}
	}
	return nil
}

func (t *Task) SetDeadline(deadline time.Time) {
	t.setTagWithPrefix(constants.DeadlineTagPrefix, deadline.Format(constants.DateFormat))
}

func (t *Task) RemoveDeadline() {
	t.setTagWithPrefix(constants.DeadlineTagPrefix, "")
}

func (t *Task) GetWaitReason() string {
	if value, ok := t.getTagWithPrefix(constants.WaitForTagPrefix); ok {
		return value
	}
	return ""
}

func (t *Task) SetWaitReason(reason string) {
	t.setTagWithPrefix(constants.WaitForTagPrefix, reason)
}

func (t *Task) IsManagedTask() bool {
	return t.hasTag(constants.TagPrefix)
}

func (t *Task) GetReminder() *time.Time {
	if value, ok := t.getTagWithPrefix(constants.ReminderTagPrefix); ok {
		// Try parsing with time first
		if reminder, err := time.Parse("2006-01-02T15:04", value); err == nil {
			return &reminder
		}
		// Fall back to date only
		if reminder, err := time.Parse(constants.DateFormat, value); err == nil {
			return &reminder
		}
	}
	return nil
}

func (t *Task) SetReminder(reminder time.Time) {
	t.setTagWithPrefix(constants.ReminderTagPrefix, reminder.Format("2006-01-02T15:04"))
}

func (t *Task) RemoveReminder() {
	t.setTagWithPrefix(constants.ReminderTagPrefix, "")
}

// Parent-child relationship methods

// GetParentID returns the parent task ID if this task is a subtask
func (t *Task) GetParentID() string {
	if value, ok := t.getTagWithPrefix(constants.ParentTagPrefix); ok {
		return value
	}
	return ""
}

// SetParentID sets the parent task ID, making this task a subtask
func (t *Task) SetParentID(parentID string) {
	t.setTagWithPrefix(constants.ParentTagPrefix, parentID)
}

// RemoveParent removes the parent relationship
func (t *Task) RemoveParent() {
	t.setTagWithPrefix(constants.ParentTagPrefix, "")
}

// HasParent returns true if this task is a subtask
func (t *Task) HasParent() bool {
	return t.GetParentID() != ""
}

// IsParentOf returns true if this task is the parent of the given task ID
func (t *Task) IsParentOf(childID string) bool {
	return t.ID == childID
}