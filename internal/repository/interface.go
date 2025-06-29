package repository

import (
	"github.com/tkancf/mdtask/internal/task"
)

// Repository defines the interface for task storage operations
type Repository interface {
	// Basic CRUD operations
	FindAll() ([]*task.Task, error)
	FindByID(id string) (*task.Task, error)
	FindByIDWithPath(id string) (*task.Task, string, error)
	Create(t *task.Task) (string, error)
	Update(t *task.Task) error
	Save(t *task.Task, filePath string) error

	// Query operations
	FindByStatus(status task.Status) ([]*task.Task, error)
	FindActive() ([]*task.Task, error)
	Search(query string) ([]*task.Task, error)
	SearchByTags(includeTags, excludeTags []string, orMode bool) ([]*task.Task, error)
}