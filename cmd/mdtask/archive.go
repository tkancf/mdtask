package mdtask

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/pkg/markdown"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [task-id]",
	Short: "Archive a task",
	Long:  `Archive a task by adding the mdtask/archived tag.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	task, err := repo.FindByID(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.IsArchived() {
		fmt.Printf("Task %s is already archived.\n", taskID)
		return nil
	}

	task.Archive()
	task.Updated = time.Now()

	var taskFilePath string
	// Try to find file by timestamp (task/YYYYMMDDHHMMSS -> YYYYMMDDHHMMSS.md)
	timestamp := strings.TrimPrefix(taskID, "task/")
	fileName := fmt.Sprintf("%s.md", timestamp)
	for _, root := range paths {
		testPath := filepath.Join(root, fileName)
		if _, err := os.Stat(testPath); err == nil {
			taskFilePath = testPath
			break
		}
	}

	if taskFilePath == "" {
		err := filepath.Walk(paths[0], func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			t, err := markdown.ParseTaskFile(content)
			if err != nil {
				return nil
			}

			if t != nil && t.ID == taskID {
				taskFilePath = path
				return filepath.SkipAll
			}
			
			return nil
		})
		
		if err != nil && err != filepath.SkipAll {
			return fmt.Errorf("failed to find task file: %w", err)
		}
	}

	if taskFilePath == "" {
		return fmt.Errorf("task file not found for: %s", taskID)
	}

	if err := repo.Save(task, taskFilePath); err != nil {
		return fmt.Errorf("failed to save archived task: %w", err)
	}

	fmt.Printf("Task %s archived successfully.\n", taskID)
	return nil
}