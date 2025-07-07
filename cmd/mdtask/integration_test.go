package mdtask

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/tkancf/mdtask/internal/output"
)

// TestContext provides a test environment for integration tests
type TestContext struct {
	t       *testing.T
	tempDir string
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
}

// NewTestContext creates a new test context with temporary directory
func NewTestContext(t *testing.T) *TestContext {
	tempDir, err := os.MkdirTemp("", "mdtask-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	return &TestContext{
		t:       t,
		tempDir: tempDir,
		stdout:  new(bytes.Buffer),
		stderr:  new(bytes.Buffer),
	}
}

// Cleanup removes temporary directory
func (tc *TestContext) Cleanup() {
	os.RemoveAll(tc.tempDir)
}

// Execute runs a command with arguments
func (tc *TestContext) Execute(args ...string) error {
	tc.stdout.Reset()
	tc.stderr.Reset()

	// Create new root command for each test
	cmd := rootCmd
	cmd.SetOut(tc.stdout)
	cmd.SetErr(tc.stderr)
	
	// Add temp dir to args
	fullArgs := append([]string{"--paths", tc.tempDir}, args...)
	cmd.SetArgs(fullArgs)

	return cmd.Execute()
}

// ExecuteWithFormat runs a command with JSON format
func (tc *TestContext) ExecuteWithFormat(format string, args ...string) error {
	fullArgs := append([]string{"--format", format}, args...)
	return tc.Execute(fullArgs...)
}

// GetStdout returns stdout content
func (tc *TestContext) GetStdout() string {
	return tc.stdout.String()
}

// GetStderr returns stderr content
func (tc *TestContext) GetStderr() string {
	return tc.stderr.String()
}

// ParseJSONOutput parses stdout as JSON
func (tc *TestContext) ParseJSONOutput(v interface{}) error {
	output := tc.stdout.String()
	
	// Find JSON object in the output
	start := strings.Index(output, "{")
	if start == -1 {
		return json.Unmarshal(tc.stdout.Bytes(), v)
	}
	
	// Find the matching closing brace
	braceCount := 0
	end := start
	for i, char := range output[start:] {
		if char == '{' {
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 {
				end = start + i + 1
				break
			}
		}
	}
	
	if braceCount != 0 {
		return json.Unmarshal(tc.stdout.Bytes(), v)
	}
	
	jsonStr := output[start:end]
	err := json.Unmarshal([]byte(jsonStr), v)
	if err != nil {
		// Debug output
		tc.t.Logf("Failed to parse JSON: %v\nExtracted JSON:\n%s\nFull output:\n%s", err, jsonStr, output)
	}
	return err
}

func TestIntegration_CreateAndListTasks(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create a task
	err := tc.Execute("new", "--title", "Integration Test Task", "--description", "Test Description", "--status", "TODO", "--content", "Test content")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	output := tc.GetStdout()
	if !strings.Contains(output, "Task created successfully!") {
		t.Error("expected success message")
	}

	// List tasks
	err = tc.Execute("list")
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	output = tc.GetStdout()
	if !strings.Contains(output, "Integration Test Task") {
		t.Error("created task not found in list")
	}
}

func TestIntegration_CreateAndGetTask_JSON(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create a task with JSON output
	err := tc.ExecuteWithFormat("json", "new", "--title", "JSON Task", "--description", "JSON Description", "--content", "JSON content")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	var created output.TaskJSON
	if err := tc.ParseJSONOutput(&created); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if created.Title != "JSON Task" {
		t.Errorf("expected title 'JSON Task', got %q", created.Title)
	}

	// Get the task with JSON output
	err = tc.ExecuteWithFormat("json", "get", created.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	var retrieved output.TaskJSON
	if err := tc.ParseJSONOutput(&retrieved); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, retrieved.ID)
	}
}

func TestIntegration_EditTask(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create a task
	err := tc.ExecuteWithFormat("json", "new", "--title", "Original Title", "--content", "Original content")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	var created output.TaskJSON
	if err := tc.ParseJSONOutput(&created); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Edit the task
	err = tc.Execute("edit", created.ID, "--title", "Updated Title", "--status", "WIP")
	if err != nil {
		t.Fatalf("failed to edit task: %v", err)
	}

	// Get updated task
	err = tc.ExecuteWithFormat("json", "get", created.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	var updated output.TaskJSON
	if err := tc.ParseJSONOutput(&updated); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}
	if updated.Status != "WIP" {
		t.Errorf("expected status 'WIP', got %q", updated.Status)
	}
}

func TestIntegration_ArchiveTask(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create a task
	err := tc.ExecuteWithFormat("json", "new", "--title", "Task to Archive", "--content", "Archive content")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	var created output.TaskJSON
	if err := tc.ParseJSONOutput(&created); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Archive the task
	err = tc.Execute("archive", created.ID)
	if err != nil {
		t.Fatalf("failed to archive task: %v", err)
	}

	// List active tasks (should not include archived)
	err = tc.ExecuteWithFormat("json", "list")
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	var tasks []output.TaskJSON
	if err := tc.ParseJSONOutput(&tasks); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	for _, task := range tasks {
		if task.ID == created.ID {
			t.Error("archived task should not appear in active list")
		}
	}

	// List archived tasks
	err = tc.ExecuteWithFormat("json", "list", "--archived")
	if err != nil {
		t.Fatalf("failed to list archived tasks: %v", err)
	}

	if err := tc.ParseJSONOutput(&tasks); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	found := false
	for _, task := range tasks {
		if task.ID == created.ID {
			found = true
			if !task.IsArchived {
				t.Error("task should be marked as archived")
			}
			break
		}
	}
	if !found {
		t.Error("archived task not found in archived list")
	}
}

func TestIntegration_SearchTasks(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create tasks with different attributes
	tasks := []struct {
		title string
		tags  string
	}{
		{"Bug Fix: Login Issue", "bug,high-priority"},
		{"Feature: Dashboard", "feature,low-priority"},
		{"Bug Fix: Profile Page", "bug,low-priority"},
	}

	for _, task := range tasks {
		err := tc.Execute("new", "--title", task.title, "--tags", task.tags, "--content", "Test content")
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Search by text
	err := tc.ExecuteWithFormat("json", "search", "Bug Fix")
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	var results []output.TaskJSON
	if err := tc.ParseJSONOutput(&results); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 bug fix tasks, got %d", len(results))
	}

	// Search by tags
	err = tc.ExecuteWithFormat("json", "search", "--tags", "bug,high-priority")
	if err != nil {
		t.Fatalf("failed to search by tags: %v", err)
	}

	if err := tc.ParseJSONOutput(&results); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 high-priority bug, got %d", len(results))
	}
}

func TestIntegration_ParentChildTasks(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	// Create parent task
	err := tc.ExecuteWithFormat("json", "new", "--title", "Parent Task", "--content", "Parent content")
	if err != nil {
		t.Fatalf("failed to create parent task: %v", err)
	}

	var parent output.TaskJSON
	if err := tc.ParseJSONOutput(&parent); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Create subtask
	err = tc.ExecuteWithFormat("json", "new", "--title", "Subtask 1", "--parent", parent.ID, "--content", "Subtask content")
	if err != nil {
		t.Fatalf("failed to create subtask: %v", err)
	}

	var subtask output.TaskJSON
	if err := tc.ParseJSONOutput(&subtask); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if subtask.ParentID != parent.ID {
		t.Errorf("expected parent ID %q, got %q", parent.ID, subtask.ParentID)
	}

	// List subtasks of parent
	err = tc.ExecuteWithFormat("json", "list", "--parent", parent.ID)
	if err != nil {
		t.Fatalf("failed to list subtasks: %v", err)
	}

	var subtasks []output.TaskJSON
	if err := tc.ParseJSONOutput(&subtasks); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(subtasks) != 1 {
		t.Errorf("expected 1 subtask, got %d", len(subtasks))
	}
}

func TestIntegration_TaskValidation(t *testing.T) {
	tc := NewTestContext(t)
	defer tc.Cleanup()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "empty title",
			args:    []string{"new", "--title", ""},
			wantErr: true,
		},
		{
			name:    "invalid status",
			args:    []string{"new", "--title", "Test", "--status", "INVALID"},
			wantErr: true,
		},
		{
			name:    "invalid deadline format",
			args:    []string{"new", "--title", "Test", "--deadline", "01/01/2024"},
			wantErr: true,
		},
		{
			name:    "valid task",
			args:    []string{"new", "--title", "Valid Task", "--status", "TODO", "--deadline", "2024-12-31", "--content", "Valid content"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tc.Execute(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}