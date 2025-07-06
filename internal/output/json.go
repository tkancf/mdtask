package output

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

// TaskJSON represents a task in JSON format
type TaskJSON struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Reminder    *time.Time `json:"reminder,omitempty"`
	IsArchived  bool       `json:"is_archived"`
	Content     string     `json:"content,omitempty"`
	ParentID    string     `json:"parent_id,omitempty"`
	FilePath    string     `json:"file_path,omitempty"`
}

// NewTaskJSON creates a TaskJSON from a task.Task
func NewTaskJSON(t *task.Task) TaskJSON {
	return TaskJSON{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.GetStatus()),
		Tags:        t.Tags,
		Created:     t.Created,
		Updated:     t.Updated,
		Deadline:    t.GetDeadline(),
		Reminder:    t.GetReminder(),
		IsArchived:  t.IsArchived(),
		Content:     t.Content,
		ParentID:    t.GetParentID(),
	}
}

// NewTaskJSONWithPath creates a TaskJSON with file path information
func NewTaskJSONWithPath(t *task.Task, filePath string) TaskJSON {
	tj := NewTaskJSON(t)
	tj.FilePath = filePath
	return tj
}

// JSONPrinter handles JSON output for tasks
type JSONPrinter struct {
	writer io.Writer
}

// NewJSONPrinter creates a new JSON printer
func NewJSONPrinter(w io.Writer) *JSONPrinter {
	if w == nil {
		w = os.Stdout
	}
	return &JSONPrinter{writer: w}
}

// PrintTask outputs a single task as JSON
func (p *JSONPrinter) PrintTask(t *task.Task) error {
	tj := NewTaskJSON(t)
	return p.printJSON(tj)
}

// PrintTaskWithPath outputs a single task with file path as JSON
func (p *JSONPrinter) PrintTaskWithPath(t *task.Task, filePath string) error {
	tj := NewTaskJSONWithPath(t, filePath)
	return p.printJSON(tj)
}

// PrintTasks outputs multiple tasks as JSON array
func (p *JSONPrinter) PrintTasks(tasks []*task.Task) error {
	jsonTasks := make([]TaskJSON, len(tasks))
	for i, t := range tasks {
		jsonTasks[i] = NewTaskJSON(t)
	}
	return p.printJSON(jsonTasks)
}

// PrintEmpty outputs an empty JSON array
func (p *JSONPrinter) PrintEmpty() error {
	return p.printJSON([]TaskJSON{})
}

// printJSON is a helper to encode and output JSON
func (p *JSONPrinter) printJSON(v interface{}) error {
	encoder := json.NewEncoder(p.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}