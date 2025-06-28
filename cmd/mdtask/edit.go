package mdtask

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/repository"
)

var editCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit an existing task",
	Long:  `Edit an existing task in your default editor.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)


	var taskFilePath string
	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}

			t, err := repo.FindByID(taskID)
			if err == nil && t != nil {
				taskFilePath = path
				return filepath.SkipAll
			}
			
			return nil
		})
		
		if taskFilePath != "" {
			break
		}
		
		if err != nil && err != filepath.SkipAll {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	}

	if taskFilePath == "" {
		fileName := fmt.Sprintf("task_%s.md", taskID[5:])
		for _, root := range paths {
			testPath := filepath.Join(root, fileName)
			if _, err := os.Stat(testPath); err == nil {
				taskFilePath = testPath
				break
			}
		}
	}

	if taskFilePath == "" {
		return fmt.Errorf("task not found: %s", taskID)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editCmd := exec.Command(editor, taskFilePath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("Task %s edited successfully.\n", taskID)
	return nil
}