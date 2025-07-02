package mdtask

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
)

var (
	editTitle       string
	editDescription string
	editStatus      string
	editTags        string
	editContent     string
	editDeadline    string
)

var editCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit an existing task",
	Long:  `Edit an existing task in your default editor or update specific fields using flags.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
	
	// Add flags for programmatic editing
	editCmd.Flags().StringVar(&editTitle, "title", "", "Update task title")
	editCmd.Flags().StringVar(&editDescription, "description", "", "Update task description")
	editCmd.Flags().StringVar(&editStatus, "status", "", "Update task status (TODO, WIP, WAIT, SCHE, DONE)")
	editCmd.Flags().StringVar(&editTags, "tags", "", "Update task tags (comma-separated)")
	editCmd.Flags().StringVar(&editContent, "content", "", "Update task content")
	editCmd.Flags().StringVar(&editDeadline, "deadline", "", "Update task deadline (YYYY-MM-DD)")
}

func runEdit(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)
	
	// Check if any flags are provided for programmatic editing
	hasFlags := editTitle != "" || editDescription != "" || editStatus != "" || 
		editTags != "" || editContent != "" || editDeadline != ""
	
	if hasFlags {
		// Programmatic editing mode
		t, err := repo.FindByID(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %s", taskID)
		}
		
		// Update fields based on flags
		if editTitle != "" {
			t.Title = editTitle
		}
		
		if cmd.Flags().Changed("description") {
			t.Description = editDescription
		}
		
		if editStatus != "" {
			switch strings.ToUpper(editStatus) {
			case "TODO":
				t.SetStatus(task.StatusTODO)
			case "WIP":
				t.SetStatus(task.StatusWIP)
			case "WAIT":
				t.SetStatus(task.StatusWAIT)
			case "SCHE":
				t.SetStatus(task.StatusSCHE)
			case "DONE":
				t.SetStatus(task.StatusDONE)
			default:
				return fmt.Errorf("invalid status: %s", editStatus)
			}
		}
		
		if editTags != "" {
			// Preserve system tags
			var systemTags []string
			for _, tag := range t.Tags {
				if tag == "mdtask" || strings.HasPrefix(tag, "mdtask/") {
					systemTags = append(systemTags, tag)
				}
			}
			
			// Parse new user tags
			userTags := strings.Split(editTags, ",")
			for i, tag := range userTags {
				userTags[i] = strings.TrimSpace(tag)
			}
			
			// Combine system and user tags
			t.Tags = append(systemTags, userTags...)
		}
		
		if cmd.Flags().Changed("content") {
			t.Content = editContent
		}
		
		if editDeadline != "" {
			if editDeadline == "none" || editDeadline == "clear" {
				// Clear deadline
				var newTags []string
				for _, tag := range t.Tags {
					if !strings.HasPrefix(tag, "mdtask/deadline/") {
						newTags = append(newTags, tag)
					}
				}
				t.Tags = newTags
			} else {
				// Parse and set deadline
				deadline, err := time.Parse("2006-01-02", editDeadline)
				if err != nil {
					return fmt.Errorf("invalid deadline format (use YYYY-MM-DD): %w", err)
				}
				t.SetDeadline(deadline)
			}
		}
		
		// Update the task
		t.Updated = time.Now()
		if err := repo.Update(t); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
		
		fmt.Printf("Task %s updated successfully.\n", taskID)
		return nil
	}
	
	// Original editor mode
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