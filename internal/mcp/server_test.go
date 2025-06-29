package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/errors"
	"github.com/tkancf/mdtask/internal/task"
)

type mockRepository struct {
	tasks map[string]*task.Task
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		tasks: make(map[string]*task.Task),
	}
}

func (m *mockRepository) Create(t *task.Task) (string, error) {
	// Generate unique ID with microseconds to avoid collisions in tests
	t.ID = fmt.Sprintf("task/%s%d", time.Now().Format("20060102150405"), time.Now().Nanosecond())
	m.tasks[t.ID] = t
	return t.ID, nil
}

func (m *mockRepository) FindByID(id string) (*task.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, errors.NotFound("task", id)
	}
	return t, nil
}

func (m *mockRepository) FindByIDWithPath(id string) (*task.Task, string, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, "", errors.NotFound("task", id)
	}
	return t, "", nil
}

func (m *mockRepository) Update(t *task.Task) error {
	m.tasks[t.ID] = t
	return nil
}

func (m *mockRepository) Save(t *task.Task, filePath string) error {
	m.tasks[t.ID] = t
	return nil
}

func (m *mockRepository) FindAll() ([]*task.Task, error) {
	tasks := make([]*task.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (m *mockRepository) FindByStatus(status task.Status) ([]*task.Task, error) {
	tasks := make([]*task.Task, 0)
	for _, t := range m.tasks {
		// Don't include archived tasks unless specifically looking for them
		if !t.IsArchived() && t.GetStatus() == status {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (m *mockRepository) FindActive() ([]*task.Task, error) {
	tasks := make([]*task.Task, 0)
	for _, t := range m.tasks {
		if !t.IsArchived() {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (m *mockRepository) Search(query string) ([]*task.Task, error) {
	tasks := make([]*task.Task, 0)
	for _, t := range m.tasks {
		if strings.Contains(strings.ToLower(t.Title), strings.ToLower(query)) || 
		   strings.Contains(strings.ToLower(t.Description), strings.ToLower(query)) || 
		   strings.Contains(strings.ToLower(t.Content), strings.ToLower(query)) {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (m *mockRepository) SearchByTags(includeTags, excludeTags []string, orMode bool) ([]*task.Task, error) {
	tasks := make([]*task.Task, 0)
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}


func TestListTasksHandler(t *testing.T) {
	repo := newMockRepository()
	cfg := config.DefaultConfig()
	server := NewServer(repo, cfg)

	// Add test tasks
	task1 := &task.Task{
		Title:   "Test Task 1",
		Tags:    []string{"mdtask"},
		Created: time.Now(),
		Updated: time.Now(),
	}
	task1.SetStatus(task.StatusTODO)
	
	task2 := &task.Task{
		Title:   "Test Task 2",
		Tags:    []string{"mdtask"},
		Created: time.Now(),
		Updated: time.Now(),
	}
	task2.SetStatus(task.StatusWIP)
	
	task3 := &task.Task{
		Title:   "Archived Task",
		Tags:    []string{"mdtask"},
		Created: time.Now(),
		Updated: time.Now(),
	}
	task3.SetStatus(task.StatusDONE)
	task3.Archive()

	repo.Create(task1)
	repo.Create(task2)
	repo.Create(task3)

	tests := []struct {
		name            string
		args            map[string]interface{}
		expectedCount   string
		shouldContain   []string
		shouldNotContain []string
	}{
		{
			name:          "List all active tasks",
			args:          map[string]interface{}{},
			expectedCount: "Found 2 tasks",
			shouldContain: []string{"Test Task 1", "Test Task 2"},
			shouldNotContain: []string{"Archived Task"},
		},
		{
			name:          "List TODO tasks",
			args:          map[string]interface{}{"status": "TODO"},
			expectedCount: "Found 1 tasks",
			shouldContain: []string{"Test Task 1"},
			shouldNotContain: []string{"Test Task 2", "Archived Task"},
		},
		{
			name:          "List with archived",
			args:          map[string]interface{}{"archived": true},
			expectedCount: "Found 3 tasks",
			shouldContain: []string{"Test Task 1", "Test Task 2", "Archived Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := server.listTasksHandler(context.Background(), request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Content) == 0 {
				t.Fatal("expected content in result")
			}
			var content string
			if len(result.Content) > 0 {
				if tc, ok := result.Content[0].(mcp.TextContent); ok {
					content = tc.Text
				} else if tc, ok := result.Content[0].(*mcp.TextContent); ok {
					content = tc.Text
				}
			}

			// Check count
			if !strings.Contains(content, tt.expectedCount) {
				t.Errorf("expected content to contain %q, got %q", tt.expectedCount, content)
			}

			// Check should contain
			for _, s := range tt.shouldContain {
				if !strings.Contains(content, s) {
					t.Errorf("expected content to contain %q", s)
				}
			}

			// Check should not contain
			for _, s := range tt.shouldNotContain {
				if strings.Contains(content, s) {
					t.Errorf("expected content NOT to contain %q", s)
				}
			}
		})
	}
}

func TestCreateTaskHandler(t *testing.T) {
	repo := newMockRepository()
	cfg := config.DefaultConfig()
	server := NewServer(repo, cfg)

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		checkTask   func(*task.Task) error
	}{
		{
			name: "Create basic task",
			args: map[string]interface{}{
				"title": "New Task",
			},
			checkTask: func(t *task.Task) error {
				if t.Title != "New Task" {
					return fmt.Errorf("expected title 'New Task', got %s", t.Title)
				}
				if !t.IsManagedTask() {
					return fmt.Errorf("task is not managed")
				}
				if t.GetStatus() != task.StatusTODO {
					return fmt.Errorf("expected status TODO, got %s", t.GetStatus())
				}
				return nil
			},
		},
		{
			name: "Create task with status",
			args: map[string]interface{}{
				"title":  "WIP Task",
				"status": "WIP",
			},
			checkTask: func(t *task.Task) error {
				if t.GetStatus() != task.StatusWIP {
					return fmt.Errorf("expected status WIP, got %s", t.GetStatus())
				}
				return nil
			},
		},
		{
			name: "Create task with tags",
			args: map[string]interface{}{
				"title": "Tagged Task",
				"tags":  []interface{}{"priority/high", "project/test"},
			},
			checkTask: func(t *task.Task) error {
				hasHigh := false
				hasTest := false
				for _, tag := range t.Tags {
					if tag == "priority/high" {
						hasHigh = true
					}
					if tag == "project/test" {
						hasTest = true
					}
				}
				if !hasHigh {
					return fmt.Errorf("expected tag priority/high")
				}
				if !hasTest {
					return fmt.Errorf("expected tag project/test")
				}
				return nil
			},
		},
		{
			name:        "Create without title",
			args:        map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := server.createTaskHandler(context.Background(), request)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check the task was created
			tasks, _ := repo.FindAll()
			if len(tasks) == 0 {
				t.Fatal("no task was created")
			}

			lastTask := tasks[len(tasks)-1]
			if tt.checkTask != nil {
				if err := tt.checkTask(lastTask); err != nil {
					t.Error(err)
				}
			}

			// Check result message
			if len(result.Content) == 0 {
				t.Fatal("expected content in result")
			}
			var content string
			if len(result.Content) > 0 {
				if tc, ok := result.Content[0].(mcp.TextContent); ok {
					content = tc.Text
				} else if tc, ok := result.Content[0].(*mcp.TextContent); ok {
					content = tc.Text
				}
			}
			if !strings.Contains(content, "Task created successfully") {
				t.Errorf("expected success message in result: %s", content)
			}
		})
	}
}

func TestUpdateTaskHandler(t *testing.T) {
	repo := newMockRepository()
	cfg := config.DefaultConfig()
	server := NewServer(repo, cfg)

	// Create initial task
	initialTask := &task.Task{
		Title:       "Original Title",
		Description: "Original Description",
		Tags:        []string{"mdtask", "mdtask/status/TODO", "keep-this"},
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	repo.Create(initialTask)

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		checkTask   func(*task.Task) error
	}{
		{
			name: "Update title",
			args: map[string]interface{}{
				"id":    initialTask.ID,
				"title": "Updated Title",
			},
			checkTask: func(t *task.Task) error {
				if t.Title != "Updated Title" {
					return fmt.Errorf("expected title 'Updated Title', got %s", t.Title)
				}
				return nil
			},
		},
		{
			name: "Update status",
			args: map[string]interface{}{
				"id":     initialTask.ID,
				"status": "DONE",
			},
			checkTask: func(t *task.Task) error {
				if t.GetStatus() != task.StatusDONE {
					return fmt.Errorf("expected status DONE, got %s", t.GetStatus())
				}
				return nil
			},
		},
		{
			name: "Add and remove tags",
			args: map[string]interface{}{
				"id":          initialTask.ID,
				"add_tags":    []interface{}{"new-tag"},
				"remove_tags": []interface{}{"keep-this"},
			},
			checkTask: func(t *task.Task) error {
				hasNewTag := false
				hasKeepThis := false
				for _, tag := range t.Tags {
					if tag == "new-tag" {
						hasNewTag = true
					}
					if tag == "keep-this" {
						hasKeepThis = true
					}
				}
				if !hasNewTag {
					return fmt.Errorf("expected tag new-tag")
				}
				if hasKeepThis {
					return fmt.Errorf("tag keep-this should have been removed")
				}
				return nil
			},
		},
		{
			name:        "Update without ID",
			args:        map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := server.updateTaskHandler(context.Background(), request)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check the task was updated
			updatedTask, _ := repo.FindByID(initialTask.ID)
			if tt.checkTask != nil {
				if err := tt.checkTask(updatedTask); err != nil {
					t.Error(err)
				}
			}

			// Check result message
			if len(result.Content) == 0 {
				t.Fatal("expected content in result")
			}
			var content string
			if len(result.Content) > 0 {
				if tc, ok := result.Content[0].(mcp.TextContent); ok {
					content = tc.Text
				} else if tc, ok := result.Content[0].(*mcp.TextContent); ok {
					content = tc.Text
				}
			}
			if !strings.Contains(content, "Task updated successfully") {
				t.Errorf("expected success message in result: %s", content)
			}
		})
	}
}