package service

import (
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/errors"
	"github.com/tkancf/mdtask/internal/task"
)

// MockTaskRepository is a mock implementation of repository for testing
type MockTaskRepository struct {
	tasks          map[string]*task.Task
	shouldFailFind bool
	shouldFailSave bool
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[string]*task.Task),
	}
}

func (m *MockTaskRepository) FindByID(id string) (*task.Task, error) {
	if m.shouldFailFind {
		return nil, errors.InternalError("mock find error", nil)
	}
	
	t, exists := m.tasks[id]
	if !exists {
		return nil, errors.NotFound("task", id)
	}
	return t, nil
}

func (m *MockTaskRepository) Create(t *task.Task) (string, error) {
	if m.shouldFailSave {
		return "", errors.InternalError("mock save error", nil)
	}
	
	if t.ID == "" {
		t.ID = "task/20240101120000"
	}
	m.tasks[t.ID] = t
	return "/path/to/" + t.ID + ".md", nil
}

func (m *MockTaskRepository) Update(t *task.Task) error {
	if m.shouldFailSave {
		return errors.InternalError("mock update error", nil)
	}
	
	if _, exists := m.tasks[t.ID]; !exists {
		return errors.NotFound("task", t.ID)
	}
	
	t.Updated = time.Now()
	m.tasks[t.ID] = t
	return nil
}

func (m *MockTaskRepository) FindAll() ([]*task.Task, error) {
	if m.shouldFailFind {
		return nil, errors.InternalError("mock find error", nil)
	}
	
	tasks := make([]*task.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Implement minimal methods to satisfy the interface
func (m *MockTaskRepository) FindByIDWithPath(id string) (*task.Task, string, error) {
	t, err := m.FindByID(id)
	if err != nil {
		return nil, "", err
	}
	return t, "/path/to/" + id + ".md", nil
}

func (m *MockTaskRepository) Save(t *task.Task, path string) error {
	return m.Update(t)
}

func (m *MockTaskRepository) FindByStatus(status task.Status) ([]*task.Task, error) {
	tasks, err := m.FindAll()
	if err != nil {
		return nil, err
	}
	
	var filtered []*task.Task
	for _, t := range tasks {
		if t.GetStatus() == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *MockTaskRepository) FindActive() ([]*task.Task, error) {
	tasks, err := m.FindAll()
	if err != nil {
		return nil, err
	}
	
	var active []*task.Task
	for _, t := range tasks {
		if !t.IsArchived() {
			active = append(active, t)
		}
	}
	return active, nil
}

func (m *MockTaskRepository) Search(query string) ([]*task.Task, error) {
	return m.FindAll()
}

func (m *MockTaskRepository) SearchByTags(include, exclude []string, orMode bool) ([]*task.Task, error) {
	return m.FindAll()
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateTaskParams
		config  *config.Config
		wantErr bool
		check   func(*testing.T, *task.Task)
	}{
		{
			name: "basic task creation",
			params: CreateTaskParams{
				Title:       "Test Task",
				Description: "Test Description",
				Content:     "Test Content",
			},
			config: &config.Config{
				Task: config.TaskConfig{},
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.Title != "Test Task" {
					t.Errorf("expected title 'Test Task', got %q", task.Title)
				}
				if task.GetStatus() != "TODO" {
					t.Errorf("expected status TODO, got %q", task.GetStatus())
				}
			},
		},
		{
			name: "task with config defaults",
			params: CreateTaskParams{
				Title: "Task",
			},
			config: &config.Config{
				Task: config.TaskConfig{
					TitlePrefix:         "[PREFIX] ",
					DescriptionTemplate: "Default Description",
					ContentTemplate:     "Default Content",
					DefaultStatus:       "WIP",
					DefaultTags:         []string{"project-x"},
				},
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.Title != "[PREFIX] Task" {
					t.Errorf("expected title with prefix, got %q", task.Title)
				}
				if task.Description != "Default Description" {
					t.Errorf("expected default description, got %q", task.Description)
				}
				if task.GetStatus() != "WIP" {
					t.Errorf("expected status WIP, got %q", task.GetStatus())
				}
				found := false
				for _, tag := range task.Tags {
					if tag == "project-x" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected tag 'project-x' not found")
				}
			},
		},
		{
			name: "task with parent",
			params: CreateTaskParams{
				Title:    "Subtask",
				ParentID: "task/20240101000000",
			},
			config: &config.Config{
				Task: config.TaskConfig{},
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.GetParentID() != "task/20240101000000" {
					t.Errorf("expected parent ID, got %q", task.GetParentID())
				}
			},
		},
		{
			name: "invalid title",
			params: CreateTaskParams{
				Title: "",
			},
			config: &config.Config{
				Task: config.TaskConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockTaskRepository()
			// Add parent task if needed
			if tt.params.ParentID != "" {
				parent := &task.Task{
					ID:     tt.params.ParentID,
					Title:  "Parent Task",
					Tags:   []string{"mdtask"},
				}
				parent.SetStatus(task.StatusTODO)
				repo.tasks[parent.ID] = parent
			}

			service := NewTaskService(repo, tt.config)
			
			task, _, err := service.CreateTask(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t, task)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		params  UpdateTaskParams
		setup   func(*MockTaskRepository)
		wantErr bool
		check   func(*testing.T, *task.Task)
	}{
		{
			name:   "update title",
			taskID: "task/20240101120000",
			params: UpdateTaskParams{
				Title: stringPtr("Updated Title"),
			},
			setup: func(repo *MockTaskRepository) {
				t := &task.Task{
					ID:    "task/20240101120000",
					Title: "Original Title",
					Tags:  []string{"mdtask"},
				}
				repo.tasks[t.ID] = t
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.Title != "Updated Title" {
					t.Errorf("expected updated title, got %q", task.Title)
				}
			},
		},
		{
			name:   "update status",
			taskID: "task/20240101120000",
			params: UpdateTaskParams{
				Status: statusPtr(task.StatusDONE),
			},
			setup: func(repo *MockTaskRepository) {
				t := &task.Task{
					ID:   "task/20240101120000",
					Tags: []string{"mdtask", "mdtask/status/TODO"},
				}
				t.SetStatus(task.StatusTODO)
				repo.tasks[t.ID] = t
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.GetStatus() != "DONE" {
					t.Errorf("expected status DONE, got %q", task.GetStatus())
				}
			},
		},
		{
			name:   "clear deadline",
			taskID: "task/20240101120000",
			params: UpdateTaskParams{
				ClearDeadline: true,
			},
			setup: func(repo *MockTaskRepository) {
				t := &task.Task{
					ID:   "task/20240101120000",
					Tags: []string{"mdtask", "mdtask/deadline/2024-12-31"},
				}
				repo.tasks[t.ID] = t
			},
			wantErr: false,
			check: func(t *testing.T, task *task.Task) {
				if task.GetDeadline() != nil {
					t.Error("expected deadline to be cleared")
				}
			},
		},
		{
			name:   "task not found",
			taskID: "task/nonexistent",
			params: UpdateTaskParams{
				Title: stringPtr("Title"),
			},
			setup:   func(repo *MockTaskRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockTaskRepository()
			if tt.setup != nil {
				tt.setup(repo)
			}

			service := NewTaskService(repo, &config.Config{})
			
			task, err := service.UpdateTask(tt.taskID, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t, task)
			}
		})
	}
}

func TestArchiveTask(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		setup   func(*MockTaskRepository)
		wantErr bool
		check   func(*testing.T, *MockTaskRepository)
	}{
		{
			name:   "archive simple task",
			taskID: "task/20240101120000",
			setup: func(repo *MockTaskRepository) {
				t := &task.Task{
					ID:   "task/20240101120000",
					Tags: []string{"mdtask"},
				}
				repo.tasks[t.ID] = t
			},
			wantErr: false,
			check: func(t *testing.T, repo *MockTaskRepository) {
				task := repo.tasks["task/20240101120000"]
				if !task.IsArchived() {
					t.Error("expected task to be archived")
				}
			},
		},
		{
			name:   "archive task with subtasks",
			taskID: "task/20240101120000",
			setup: func(repo *MockTaskRepository) {
				parent := &task.Task{
					ID:   "task/20240101120000",
					Tags: []string{"mdtask"},
				}
				repo.tasks[parent.ID] = parent

				child1 := &task.Task{
					ID:   "task/20240101130000",
					Tags: []string{"mdtask", "mdtask/parent/task/20240101120000"},
				}
				repo.tasks[child1.ID] = child1

				child2 := &task.Task{
					ID:   "task/20240101140000",
					Tags: []string{"mdtask", "mdtask/parent/task/20240101120000"},
				}
				repo.tasks[child2.ID] = child2
			},
			wantErr: false,
			check: func(t *testing.T, repo *MockTaskRepository) {
				// Check parent is archived
				parent := repo.tasks["task/20240101120000"]
				if !parent.IsArchived() {
					t.Error("expected parent to be archived")
				}

				// Check children are archived
				child1 := repo.tasks["task/20240101130000"]
				if !child1.IsArchived() {
					t.Error("expected child1 to be archived")
				}

				child2 := repo.tasks["task/20240101140000"]
				if !child2.IsArchived() {
					t.Error("expected child2 to be archived")
				}
			},
		},
		{
			name:   "already archived",
			taskID: "task/20240101120000",
			setup: func(repo *MockTaskRepository) {
				t := &task.Task{
					ID:   "task/20240101120000",
					Tags: []string{"mdtask", "mdtask/archived"},
				}
				repo.tasks[t.ID] = t
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockTaskRepository()
			if tt.setup != nil {
				tt.setup(repo)
			}

			service := NewTaskService(repo, &config.Config{})
			
			_, err := service.ArchiveTask(tt.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t, repo)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s task.Status) *task.Status {
	return &s
}