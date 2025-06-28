package web

import (
	"net/http"
	"time"

	"github.com/tkan/mdtask/internal/constants"
	"github.com/tkan/mdtask/internal/task"
)

// parseTaskForm extracts task data from HTTP form
func (s *Server) parseTaskForm(r *http.Request, t *task.Task) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Basic fields
	t.Title = r.FormValue("title")
	t.Description = r.FormValue("description")
	t.Content = r.FormValue("content")

	// Parse tags
	formTags := r.FormValue("tags")
	additionalTags := r.Form["additional_tags"]
	t.Tags = parseFormTags(formTags, additionalTags)

	// Ensure mdtask tag
	if !t.IsManagedTask() {
		t.Tags = append(t.Tags, constants.TagPrefix)
	}

	// Set status
	if status := r.FormValue("status"); status != "" {
		t.SetStatus(getStatusFromForm(status))
	}

	// Set deadline
	if deadline := r.FormValue("deadline"); deadline != "" {
		if d, err := time.Parse(constants.DateFormat, deadline); err == nil {
			t.SetDeadline(d)
		}
	} else {
		t.RemoveDeadline()
	}

	// Set reminder
	reminderDate := r.FormValue("reminder_date")
	reminderTime := r.FormValue("reminder_time")
	if reminderDate != "" {
		reminderStr := reminderDate
		if reminderTime != "" {
			reminderStr += "T" + reminderTime
			if reminder, err := time.Parse("2006-01-02T15:04", reminderStr); err == nil {
				t.SetReminder(reminder)
			}
		} else {
			if reminder, err := time.Parse(constants.DateFormat, reminderDate); err == nil {
				t.SetReminder(reminder)
			}
		}
	} else {
		t.RemoveReminder()
	}

	return nil
}

// applyTaskConfig applies configuration settings to a new task
func (s *Server) applyTaskConfig(t *task.Task) {
	// Apply title prefix
	if s.config.Task.TitlePrefix != "" && !hasPrefix(t.Title, s.config.Task.TitlePrefix) {
		t.Title = s.config.Task.TitlePrefix + t.Title
	}

	// Apply description template
	if t.Description == "" && s.config.Task.DescriptionTemplate != "" {
		t.Description = s.config.Task.DescriptionTemplate
	}

	// Apply content template
	if t.Content == "" && s.config.Task.ContentTemplate != "" {
		t.Content = s.config.Task.ContentTemplate
	}

	// Add default tags
	existingTags := make(map[string]bool)
	for _, tag := range t.Tags {
		existingTags[tag] = true
	}
	
	for _, tag := range s.config.Task.DefaultTags {
		if !existingTags[tag] {
			t.Tags = append(t.Tags, tag)
		}
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}