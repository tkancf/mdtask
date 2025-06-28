package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tkan/mdtask/internal/task"
	"github.com/tkan/mdtask/pkg/markdown"
)

type TaskRepository struct {
	rootPaths []string
}

func NewTaskRepository(rootPaths []string) *TaskRepository {
	return &TaskRepository{
		rootPaths: rootPaths,
	}
}

func (r *TaskRepository) FindAll() ([]*task.Task, error) {
	var tasks []*task.Task

	for _, root := range r.rootPaths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}

			t, err := r.loadTask(path)
			if err != nil {
				return nil
			}

			if t != nil && t.IsManagedTask() {
				tasks = append(tasks, t)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", root, err)
		}
	}

	return tasks, nil
}

func (r *TaskRepository) FindByID(id string) (*task.Task, error) {
	tasks, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

func (r *TaskRepository) Save(t *task.Task, filePath string) error {
	content, err := markdown.WriteTaskFile(t)
	if err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func (r *TaskRepository) Create(t *task.Task) (string, error) {
	if t.ID == "" {
		t.ID = markdown.GenerateTaskID()
	}

	if !t.IsManagedTask() {
		t.Tags = append(t.Tags, "mdtask")
	}

	if t.GetStatus() == "" {
		t.SetStatus(task.StatusTODO)
	}

	// Extract timestamp from ID (task/YYYYMMDDHHMMSS -> YYYYMMDDHHMMSS.md)
	timestamp := strings.TrimPrefix(t.ID, "task/")
	fileName := fmt.Sprintf("%s.md", timestamp)
	filePath := filepath.Join(r.rootPaths[0], fileName)

	if err := r.Save(t, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func (r *TaskRepository) loadTask(path string) (*task.Task, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	t, err := markdown.ParseTaskFile(content)
	if err != nil {
		return nil, nil
	}

	return t, nil
}

func (r *TaskRepository) FindByStatus(status task.Status) ([]*task.Task, error) {
	allTasks, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	var filtered []*task.Task
	for _, t := range allTasks {
		if t.GetStatus() == status {
			filtered = append(filtered, t)
		}
	}

	return filtered, nil
}

func (r *TaskRepository) FindActive() ([]*task.Task, error) {
	allTasks, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	var active []*task.Task
	for _, t := range allTasks {
		if !t.IsArchived() {
			active = append(active, t)
		}
	}

	return active, nil
}

func (r *TaskRepository) Search(query string) ([]*task.Task, error) {
	allTasks, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matched []*task.Task
	
	for _, t := range allTasks {
		if strings.Contains(strings.ToLower(t.Title), query) ||
			strings.Contains(strings.ToLower(t.Description), query) ||
			strings.Contains(strings.ToLower(t.Content), query) {
			matched = append(matched, t)
			continue
		}

		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matched = append(matched, t)
				break
			}
		}
	}

	return matched, nil
}