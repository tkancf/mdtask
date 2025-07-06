package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

func TestNewTaskJSON(t *testing.T) {
	deadline := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	reminder := time.Date(2024, 12, 25, 14, 30, 0, 0, time.UTC)
	
	testTask := &task.Task{
		ID:          "task/20240101120000",
		Title:       "Test Task",
		Description: "Test Description",
		Tags:        []string{"mdtask", "mdtask/status/TODO", "test-tag"},
		Created:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
		Content:     "Test Content",
	}
	testTask.SetStatus(task.StatusTODO)
	testTask.SetDeadline(deadline)
	testTask.SetReminder(reminder)
	testTask.SetParentID("task/20240101000000")

	tj := NewTaskJSON(testTask)

	if tj.ID != testTask.ID {
		t.Errorf("expected ID %q, got %q", testTask.ID, tj.ID)
	}
	if tj.Title != testTask.Title {
		t.Errorf("expected Title %q, got %q", testTask.Title, tj.Title)
	}
	if tj.Status != "TODO" {
		t.Errorf("expected Status TODO, got %q", tj.Status)
	}
	if tj.ParentID != "task/20240101000000" {
		t.Errorf("expected ParentID %q, got %q", "task/20240101000000", tj.ParentID)
	}
	if tj.Deadline == nil || !tj.Deadline.Equal(deadline) {
		t.Error("deadline not set correctly")
	}
	if tj.Reminder == nil || !tj.Reminder.Equal(reminder) {
		t.Error("reminder not set correctly")
	}
}

func TestNewTaskJSONWithPath(t *testing.T) {
	testTask := &task.Task{
		ID:    "task/20240101120000",
		Title: "Test Task",
	}
	
	filePath := "/path/to/task.md"
	tj := NewTaskJSONWithPath(testTask, filePath)

	if tj.FilePath != filePath {
		t.Errorf("expected FilePath %q, got %q", filePath, tj.FilePath)
	}
}

func TestJSONPrinter_PrintTask(t *testing.T) {
	var buf bytes.Buffer
	printer := NewJSONPrinter(&buf)

	testTask := &task.Task{
		ID:          "task/20240101120000",
		Title:       "Test Task",
		Description: "Test Description",
		Tags:        []string{"mdtask", "mdtask/status/TODO"},
		Created:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	testTask.SetStatus(task.StatusTODO)

	err := printer.PrintTask(testTask)
	if err != nil {
		t.Fatalf("PrintTask() error = %v", err)
	}

	var result TaskJSON
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result.ID != testTask.ID {
		t.Errorf("expected ID %q, got %q", testTask.ID, result.ID)
	}
	if result.Title != testTask.Title {
		t.Errorf("expected Title %q, got %q", testTask.Title, result.Title)
	}
}

func TestJSONPrinter_PrintTasks(t *testing.T) {
	var buf bytes.Buffer
	printer := NewJSONPrinter(&buf)

	tasks := []*task.Task{
		{
			ID:    "task/20240101120000",
			Title: "Task 1",
			Tags:  []string{"mdtask"},
		},
		{
			ID:    "task/20240101130000",
			Title: "Task 2",
			Tags:  []string{"mdtask"},
		},
	}

	err := printer.PrintTasks(tasks)
	if err != nil {
		t.Fatalf("PrintTasks() error = %v", err)
	}

	var results []TaskJSON
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(results))
	}

	if results[0].ID != tasks[0].ID {
		t.Errorf("expected first task ID %q, got %q", tasks[0].ID, results[0].ID)
	}
	if results[1].ID != tasks[1].ID {
		t.Errorf("expected second task ID %q, got %q", tasks[1].ID, results[1].ID)
	}
}

func TestJSONPrinter_PrintEmpty(t *testing.T) {
	var buf bytes.Buffer
	printer := NewJSONPrinter(&buf)

	err := printer.PrintEmpty()
	if err != nil {
		t.Fatalf("PrintEmpty() error = %v", err)
	}

	var results []TaskJSON
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected empty array, got %d items", len(results))
	}
}

func TestJSONPrinter_NilWriter(t *testing.T) {
	printer := NewJSONPrinter(nil)
	
	// Should not panic and should write to stdout
	testTask := &task.Task{
		ID:    "task/20240101120000",
		Title: "Test Task",
		Tags:  []string{"mdtask"},
	}
	
	// These calls should not panic
	_ = printer.PrintTask(testTask)
	_ = printer.PrintTasks([]*task.Task{testTask})
	_ = printer.PrintEmpty()
}