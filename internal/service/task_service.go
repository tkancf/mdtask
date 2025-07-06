package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/errors"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
)

// TaskService handles business logic for task operations
type TaskService struct {
	repo   *repository.TaskRepository
	config *config.Config
}

// NewTaskService creates a new task service
func NewTaskService(repo *repository.TaskRepository, cfg *config.Config) *TaskService {
	return &TaskService{
		repo:   repo,
		config: cfg,
	}
}

// CreateTask creates a new task with the given parameters
func (s *TaskService) CreateTask(params CreateTaskParams) (*task.Task, string, error) {
	// Apply title prefix from config
	title := params.Title
	if s.config.Task.TitlePrefix != "" {
		title = s.config.Task.TitlePrefix + title
	}

	// Validate title
	if err := task.ValidateTitle(title); err != nil {
		return nil, "", errors.InvalidInput("title", title)
	}

	// Use description template if not provided
	description := params.Description
	if description == "" && s.config.Task.DescriptionTemplate != "" {
		description = s.config.Task.DescriptionTemplate
	}

	// Validate description
	if err := task.ValidateDescription(description); err != nil {
		return nil, "", errors.InvalidInput("description", description)
	}

	// Use content template if not provided
	content := params.Content
	if content == "" && s.config.Task.ContentTemplate != "" {
		content = s.config.Task.ContentTemplate
	}

	// Merge tags
	allTags := []string{"mdtask"}
	allTags = append(allTags, s.config.Task.DefaultTags...)
	allTags = append(allTags, params.Tags...)

	now := time.Now()
	t := &task.Task{
		Title:       title,
		Description: description,
		Created:     now,
		Updated:     now,
		Content:     content,
		Tags:        allTags,
		Aliases:     []string{},
	}

	// Set status
	status := params.Status
	if status == "" {
		status = s.config.Task.DefaultStatus
	}
	if status == "" {
		status = string(task.StatusTODO)
	}
	t.SetStatus(task.Status(status))

	// Set deadline
	if params.Deadline != nil {
		t.SetDeadline(*params.Deadline)
	}

	// Set reminder
	if params.Reminder != nil {
		t.SetReminder(*params.Reminder)
	}

	// Handle parent task
	if params.ParentID != "" {
		parentTask, err := s.repo.FindByID(params.ParentID)
		if err != nil {
			return nil, "", errors.NotFound("parent task", params.ParentID)
		}

		t.SetParentID(params.ParentID)

		// Optionally inherit parent's status
		if params.Status == "" && status == string(task.StatusTODO) {
			t.SetStatus(parentTask.GetStatus())
		}
	}

	// Create the task
	filePath, err := s.repo.Create(t)
	if err != nil {
		return nil, "", err
	}

	return t, filePath, nil
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(taskID string, params UpdateTaskParams) (*task.Task, error) {
	t, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	// Update title if provided
	if params.Title != nil {
		if err := task.ValidateTitle(*params.Title); err != nil {
			return nil, errors.InvalidInput("title", *params.Title)
		}
		t.Title = *params.Title
	}

	// Update description if provided
	if params.Description != nil {
		if err := task.ValidateDescription(*params.Description); err != nil {
			return nil, errors.InvalidInput("description", *params.Description)
		}
		t.Description = *params.Description
	}

	// Update status if provided
	if params.Status != nil {
		t.SetStatus(*params.Status)
	}

	// Update tags if provided
	if params.Tags != nil {
		// Preserve system tags
		var systemTags []string
		for _, tag := range t.Tags {
			if tag == "mdtask" || strings.HasPrefix(tag, "mdtask/") {
				systemTags = append(systemTags, tag)
			}
		}
		t.Tags = append(systemTags, *params.Tags...)
	}

	// Update content if provided
	if params.Content != nil {
		t.Content = *params.Content
	}

	// Update deadline if provided
	if params.ClearDeadline {
		var newTags []string
		for _, tag := range t.Tags {
			if !strings.HasPrefix(tag, "mdtask/deadline/") {
				newTags = append(newTags, tag)
			}
		}
		t.Tags = newTags
	} else if params.Deadline != nil {
		t.SetDeadline(*params.Deadline)
	}

	// Update reminder if provided
	if params.ClearReminder {
		var newTags []string
		for _, tag := range t.Tags {
			if !strings.HasPrefix(tag, "mdtask/reminder/") {
				newTags = append(newTags, tag)
			}
		}
		t.Tags = newTags
	} else if params.Reminder != nil {
		t.SetReminder(*params.Reminder)
	}

	// Save the updated task
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}

	return t, nil
}

// ArchiveTask archives a task
func (s *TaskService) ArchiveTask(taskID string) (*task.Task, error) {
	t, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	if t.IsArchived() {
		return nil, errors.InvalidInput("task", fmt.Sprintf("%s is already archived", taskID))
	}

	// Archive subtasks first if any
	if err := s.archiveSubtasks(taskID); err != nil {
		return nil, err
	}

	t.Archive()
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}

	return t, nil
}

// UnarchiveTask unarchives a task
func (s *TaskService) UnarchiveTask(taskID string) (*task.Task, error) {
	t, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	if !t.IsArchived() {
		return nil, errors.InvalidInput("task", fmt.Sprintf("%s is not archived", taskID))
	}

	// Check if parent is archived
	if parentID := t.GetParentID(); parentID != "" {
		parent, err := s.repo.FindByID(parentID)
		if err == nil && parent.IsArchived() {
			return nil, errors.InvalidInput("task", "cannot unarchive subtask when parent is archived")
		}
	}

	t.Unarchive()
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}

	return t, nil
}

// GetTaskWithSubtasks returns a task with all its subtasks
func (s *TaskService) GetTaskWithSubtasks(taskID string) (*task.Task, []*task.Task, error) {
	parent, err := s.repo.FindByID(taskID)
	if err != nil {
		return nil, nil, err
	}

	subtasks, err := s.findSubtasks(taskID)
	if err != nil {
		return nil, nil, err
	}

	return parent, subtasks, nil
}

// Helper method to find all subtasks
func (s *TaskService) findSubtasks(parentID string) ([]*task.Task, error) {
	allTasks, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var subtasks []*task.Task
	for _, t := range allTasks {
		if t.GetParentID() == parentID {
			subtasks = append(subtasks, t)
		}
	}

	return subtasks, nil
}

// Helper method to archive all subtasks
func (s *TaskService) archiveSubtasks(parentID string) error {
	subtasks, err := s.findSubtasks(parentID)
	if err != nil {
		return err
	}

	for _, subtask := range subtasks {
		if !subtask.IsArchived() {
			// Recursively archive subtasks
			if err := s.archiveSubtasks(subtask.ID); err != nil {
				return err
			}
			
			subtask.Archive()
			if err := s.repo.Update(subtask); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateTaskParams holds parameters for creating a task
type CreateTaskParams struct {
	Title       string
	Description string
	Content     string
	Tags        []string
	Status      string
	Deadline    *time.Time
	Reminder    *time.Time
	ParentID    string
}

// UpdateTaskParams holds parameters for updating a task
type UpdateTaskParams struct {
	Title         *string
	Description   *string
	Status        *task.Status
	Tags          *[]string
	Content       *string
	Deadline      *time.Time
	ClearDeadline bool
	Reminder      *time.Time
	ClearReminder bool
}