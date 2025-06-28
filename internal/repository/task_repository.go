package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// FindByIDWithPath finds a task by ID and returns the task and its file path
func (r *TaskRepository) FindByIDWithPath(id string) (*task.Task, string, error) {
	for _, root := range r.rootPaths {
		var foundTask *task.Task
		var foundPath string
		
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

			if t != nil && t.ID == id {
				foundTask = t
				foundPath = path
				return filepath.SkipDir // Stop walking
			}

			return nil
		})

		if err != nil && err != filepath.SkipDir {
			continue
		}
		
		if foundTask != nil {
			return foundTask, foundPath, nil
		}
	}

	// Try to find by filename pattern
	timestamp := strings.TrimPrefix(id, "task/")
	for _, root := range r.rootPaths {
		// Try exact match
		path := filepath.Join(root, timestamp+".md")
		if t, err := r.loadTask(path); err == nil && t != nil && t.ID == id {
			return t, path, nil
		}
		
		// Try with suffix
		for i := 1; i < 10; i++ {
			path = filepath.Join(root, fmt.Sprintf("%s_%d.md", timestamp, i))
			if t, err := r.loadTask(path); err == nil && t != nil && t.ID == id {
				return t, path, nil
			}
		}
	}

	return nil, "", fmt.Errorf("task not found: %s", id)
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
	baseFileName := timestamp
	
	// Check if file already exists and add suffix if needed
	var filePath string
	for i := 0; i < 100; i++ {
		var fileName string
		if i == 0 {
			fileName = fmt.Sprintf("%s.md", baseFileName)
		} else {
			fileName = fmt.Sprintf("%s_%d.md", baseFileName, i)
			// Update task ID to match filename
			t.ID = fmt.Sprintf("task/%s_%d", timestamp, i)
		}
		filePath = filepath.Join(r.rootPaths[0], fileName)
		
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
	}

	if err := r.Save(t, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func (r *TaskRepository) Update(t *task.Task) error {
	_, filePath, err := r.FindByIDWithPath(t.ID)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	// Update the updated timestamp
	t.Updated = time.Now()

	if err := r.Save(t, filePath); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return nil
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

// SearchByTags searches tasks by tag combinations with AND/OR logic
func (r *TaskRepository) SearchByTags(includeTags, excludeTags []string, orMode bool) ([]*task.Task, error) {
	allTasks, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	var matched []*task.Task
	
	for _, t := range allTasks {
		if t.IsArchived() {
			continue
		}
		
		// Check exclude tags first
		excluded := false
		for _, excludeTag := range excludeTags {
			if hasTag(t.Tags, excludeTag) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		
		// Check include tags
		if len(includeTags) == 0 {
			matched = append(matched, t)
			continue
		}
		
		if orMode {
			// OR mode: task must have at least one of the include tags
			for _, includeTag := range includeTags {
				if hasTag(t.Tags, includeTag) {
					matched = append(matched, t)
					break
				}
			}
		} else {
			// AND mode: task must have all include tags
			hasAll := true
			for _, includeTag := range includeTags {
				if !hasTag(t.Tags, includeTag) {
					hasAll = false
					break
				}
			}
			if hasAll {
				matched = append(matched, t)
			}
		}
	}
	
	return matched, nil
}

// hasTag checks if a tag exists in the tag list (case-insensitive)
func hasTag(tags []string, searchTag string) bool {
	searchTag = strings.ToLower(searchTag)
	for _, tag := range tags {
		if strings.ToLower(tag) == searchTag {
			return true
		}
	}
	return false
}