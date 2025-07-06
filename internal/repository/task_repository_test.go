package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

func TestTaskRepository_Create(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{
			name: "create simple task",
			task: &task.Task{
				Title:       "Test Task",
				Description: "Test Description",
				Created:     time.Now(),
				Updated:     time.Now(),
				Tags:        []string{},
			},
			wantErr: false,
		},
		{
			name: "create task with existing ID",
			task: &task.Task{
				ID:          "task/20240101120000",
				Title:       "Task with ID",
				Description: "Test Description",
				Created:     time.Now(),
				Updated:     time.Now(),
				Tags:        []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := repo.Create(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Error("task file was not created")
				}

				// Verify task has ID
				if tt.task.ID == "" {
					t.Error("task ID was not set")
				}

				// Verify task is managed
				if !tt.task.IsManagedTask() {
					t.Error("task should be managed")
				}

				// Verify default status
				if tt.task.GetStatus() == "" {
					t.Error("task status was not set")
				}
			}
		})
	}
}

func TestTaskRepository_FindByID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create a task
	originalTask := &task.Task{
		Title:       "Find Test Task",
		Description: "Test Description",
		Created:     time.Now(),
		Updated:     time.Now(),
		Content:     "Test Content",
		Tags:        []string{},
	}

	_, err = repo.Create(originalTask)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Test finding the task
	foundTask, err := repo.FindByID(originalTask.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if foundTask.ID != originalTask.ID {
		t.Errorf("expected ID %q, got %q", originalTask.ID, foundTask.ID)
	}
	if foundTask.Title != originalTask.Title {
		t.Errorf("expected title %q, got %q", originalTask.Title, foundTask.Title)
	}

	// Test finding non-existent task
	_, err = repo.FindByID("task/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent task")
	}

	// Test finding by filename pattern
	foundTask2, err := repo.FindByID(originalTask.ID)
	if err != nil {
		t.Errorf("FindByID() by filename pattern error = %v", err)
	}
	if foundTask2.ID != originalTask.ID {
		t.Error("failed to find task by filename pattern")
	}
}

func TestTaskRepository_Update(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create a task
	originalTask := &task.Task{
		Title:       "Original Title",
		Description: "Original Description",
		Created:     time.Now(),
		Updated:     time.Now(),
		Tags:        []string{},
	}

	_, err = repo.Create(originalTask)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Update the task
	originalTask.Title = "Updated Title"
	originalTask.Description = "Updated Description"
	originalTask.SetStatus(task.StatusWIP)

	err = repo.Update(originalTask)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	updatedTask, err := repo.FindByID(originalTask.ID)
	if err != nil {
		t.Fatalf("failed to find updated task: %v", err)
	}

	if updatedTask.Title != "Updated Title" {
		t.Errorf("expected title %q, got %q", "Updated Title", updatedTask.Title)
	}
	if updatedTask.Description != "Updated Description" {
		t.Errorf("expected description %q, got %q", "Updated Description", updatedTask.Description)
	}
	if updatedTask.GetStatus() != task.StatusWIP {
		t.Errorf("expected status WIP, got %q", updatedTask.GetStatus())
	}
}

func TestTaskRepository_FindAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create multiple tasks
	tasks := []struct {
		title  string
		status task.Status
	}{
		{"Task 1", task.StatusTODO},
		{"Task 2", task.StatusWIP},
		{"Task 3", task.StatusDONE},
	}

	for _, tt := range tasks {
		task := &task.Task{
			Title:   tt.title,
			Created: time.Now(),
			Updated: time.Now(),
			Tags:    []string{},
		}
		task.SetStatus(tt.status)
		_, err := repo.Create(task)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Find all tasks
	allTasks, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(allTasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(allTasks))
	}
}

func TestTaskRepository_FindByStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create tasks with different statuses
	statuses := []task.Status{
		task.StatusTODO,
		task.StatusTODO,
		task.StatusWIP,
		task.StatusDONE,
	}

	for i, status := range statuses {
		task := &task.Task{
			Title:   fmt.Sprintf("Task %d", i+1),
			Created: time.Now(),
			Updated: time.Now(),
			Tags:    []string{},
		}
		task.SetStatus(status)
		_, err := repo.Create(task)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Find TODO tasks
	todoTasks, err := repo.FindByStatus(task.StatusTODO)
	if err != nil {
		t.Fatalf("FindByStatus() error = %v", err)
	}

	if len(todoTasks) != 2 {
		t.Errorf("expected 2 TODO tasks, got %d", len(todoTasks))
	}

	// Find WIP tasks
	wipTasks, err := repo.FindByStatus(task.StatusWIP)
	if err != nil {
		t.Fatalf("FindByStatus() error = %v", err)
	}

	if len(wipTasks) != 1 {
		t.Errorf("expected 1 WIP task, got %d", len(wipTasks))
	}
}

func TestTaskRepository_FindActive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create active and archived tasks
	tasks := []struct {
		title    string
		archived bool
	}{
		{"Active Task 1", false},
		{"Active Task 2", false},
		{"Archived Task", true},
	}

	for _, tt := range tasks {
		task := &task.Task{
			Title:   tt.title,
			Created: time.Now(),
			Updated: time.Now(),
			Tags:    []string{},
		}
		if tt.archived {
			task.Archive()
		}
		_, err := repo.Create(task)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Find active tasks
	activeTasks, err := repo.FindActive()
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}

	if len(activeTasks) != 2 {
		t.Errorf("expected 2 active tasks, got %d", len(activeTasks))
	}

	// Verify no archived tasks in results
	for _, task := range activeTasks {
		if task.IsArchived() {
			t.Errorf("found archived task in active results: %s", task.Title)
		}
	}
}

func TestTaskRepository_Search(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create tasks with searchable content
	tasks := []struct {
		title       string
		description string
		content     string
		tags        []string
	}{
		{
			title:       "Bug Fix: Login Issue",
			description: "Users cannot login",
			content:     "Investigation shows database connection problem",
			tags:        []string{"bug", "urgent"},
		},
		{
			title:       "Feature: Dashboard",
			description: "Add new dashboard",
			content:     "Dashboard should show user statistics",
			tags:        []string{"feature", "enhancement"},
		},
		{
			title:       "Bug Fix: Profile Page",
			description: "Profile page crashes",
			content:     "Null pointer exception in profile controller",
			tags:        []string{"bug"},
		},
	}

	for _, tt := range tasks {
		task := &task.Task{
			Title:       tt.title,
			Description: tt.description,
			Content:     tt.content,
			Created:     time.Now(),
			Updated:     time.Now(),
			Tags:        append([]string{"mdtask"}, tt.tags...),
		}
		_, err := repo.Create(task)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	// Search tests
	tests := []struct {
		query    string
		expected int
	}{
		{"bug", 2},          // In title
		{"dashboard", 2},    // In title and content
		{"login", 1},        // In description
		{"statistics", 1},   // In content
		{"urgent", 1},       // In tags
		{"nonexistent", 0},  // Not found
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, err := repo.Search(tt.query)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("expected %d results for %q, got %d", tt.expected, tt.query, len(results))
			}
		})
	}
}

func TestTaskRepository_SearchByTags(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create tasks with different tag combinations
	tasks := []struct {
		title string
		tags  []string
	}{
		{"Task 1", []string{"type/bug", "priority/high"}},
		{"Task 2", []string{"type/feature", "priority/low"}},
		{"Task 3", []string{"type/bug", "priority/low"}},
		{"Task 4", []string{"type/bug", "priority/high", "status/done"}},
	}

	for _, tt := range tasks {
		task := &task.Task{
			Title:   tt.title,
			Created: time.Now(),
			Updated: time.Now(),
			Tags:    append([]string{"mdtask"}, tt.tags...),
		}
		_, err := repo.Create(task)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	tests := []struct {
		name        string
		includeTags []string
		excludeTags []string
		orMode      bool
		expected    int
	}{
		{
			name:        "AND mode - bug AND high priority",
			includeTags: []string{"type/bug", "priority/high"},
			orMode:      false,
			expected:    2, // Task 1 and 4
		},
		{
			name:        "OR mode - bug OR feature",
			includeTags: []string{"type/bug", "type/feature"},
			orMode:      true,
			expected:    4, // All tasks
		},
		{
			name:        "Exclude done tasks",
			includeTags: []string{"type/bug"},
			excludeTags: []string{"status/done"},
			orMode:      false,
			expected:    2, // Task 1 and 3
		},
		{
			name:        "No include tags",
			includeTags: []string{},
			excludeTags: []string{"status/done"},
			orMode:      false,
			expected:    3, // All except Task 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.SearchByTags(tt.includeTags, tt.excludeTags, tt.orMode)
			if err != nil {
				t.Fatalf("SearchByTags() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("expected %d results, got %d", tt.expected, len(results))
			}
		})
	}
}

func TestTaskRepository_FilenameSuffix(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := NewTaskRepository([]string{tempDir})

	// Create task with specific timestamp
	task1 := &task.Task{
		ID:      "task/20240101120000",
		Title:   "First Task",
		Created: time.Now(),
		Updated: time.Now(),
		Tags:    []string{},
	}

	filePath1, err := repo.Create(task1)
	if err != nil {
		t.Fatalf("failed to create first task: %v", err)
	}

	// Create another task with same timestamp (should get suffix)
	task2 := &task.Task{
		ID:      "task/20240101120000",
		Title:   "Second Task",
		Created: time.Now(),
		Updated: time.Now(),
		Tags:    []string{},
	}

	filePath2, err := repo.Create(task2)
	if err != nil {
		t.Fatalf("failed to create second task: %v", err)
	}

	// Verify different file paths
	if filePath1 == filePath2 {
		t.Error("expected different file paths for tasks with same timestamp")
	}

	// Verify second task has suffix in ID
	if task2.ID == task1.ID {
		t.Error("expected different IDs for tasks with same timestamp")
	}

	// Verify second task ID has suffix
	if task2.ID != "task/20240101120000_1" {
		t.Errorf("expected ID with suffix, got %q", task2.ID)
	}
}