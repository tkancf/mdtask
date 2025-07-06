package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/repository"
)

// Context holds common dependencies for CLI commands
type Context struct {
	Config *config.Config
	Repo   *repository.TaskRepository
	Paths  []string
}

// LoadContext loads configuration and creates repository from command flags
func LoadContext(cmd *cobra.Command) (*Context, error) {
	// Load configuration
	cfg, err := config.LoadFromDefaultLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get paths from flags or config
	paths, _ := cmd.Flags().GetStringSlice("paths")
	if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
		paths = cfg.Paths
	}

	// Create repository
	repo := repository.NewTaskRepository(paths)

	return &Context{
		Config: cfg,
		Repo:   repo,
		Paths:  paths,
	}, nil
}

// LoadContextWithPaths is similar to LoadContext but allows custom path handling
func LoadContextWithPaths(cfg *config.Config, paths []string) *Context {
	if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
		paths = cfg.Paths
	}

	repo := repository.NewTaskRepository(paths)

	return &Context{
		Config: cfg,
		Repo:   repo,
		Paths:  paths,
	}
}