package mdtask

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/errors"
	"github.com/tkancf/mdtask/internal/output"
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
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}
	
	taskID, err := cli.NormalizeTaskID(args[0])
	if err != nil {
		return err
	}
	
	// Check if any flags are provided for programmatic editing
	hasFlags := editTitle != "" || editDescription != "" || editStatus != "" || 
		editTags != "" || editContent != "" || editDeadline != ""
	
	if hasFlags {
		// Programmatic editing mode
		t, err := ctx.Repo.FindByID(taskID)
		if err != nil {
			return err
		}
		
		// Update fields based on flags
		if editTitle != "" {
			if err := task.ValidateTitle(editTitle); err != nil {
				return fmt.Errorf("invalid title: %w", err)
			}
			t.Title = editTitle
		}
		
		if cmd.Flags().Changed("description") {
			if err := task.ValidateDescription(editDescription); err != nil {
				return fmt.Errorf("invalid description: %w", err)
			}
			t.Description = editDescription
		}
		
		if editStatus != "" {
			status, err := cli.ValidateStatus(strings.ToUpper(editStatus))
			if err != nil {
				return err
			}
			t.SetStatus(status)
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
				deadline, err := cli.ParseDeadline(editDeadline)
				if err != nil {
					return err
				}
				if deadline != nil {
					t.SetDeadline(*deadline)
				}
			}
		}
		
		// Update the task
		if err := ctx.Repo.Update(t); err != nil {
			return err
		}
		
		if outputFormat == "json" {
			printer := output.NewJSONPrinter(os.Stdout)
			return printer.PrintTask(t)
		}
		
		fmt.Printf("Task %s updated successfully.\n", taskID)
		return nil
	}
	
	// Editor mode - find task file path
	_, taskFilePath, err := ctx.Repo.FindByIDWithPath(taskID)
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, taskFilePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return errors.InternalError("failed to open editor", err)
	}

	if outputFormat == "json" {
		// Reload task to get updated version
		t, err := ctx.Repo.FindByID(taskID)
		if err != nil {
			return err
		}
		printer := output.NewJSONPrinter(os.Stdout)
		return printer.PrintTask(t)
	}
	
	fmt.Printf("Task %s edited successfully.\n", taskID)
	return nil
}