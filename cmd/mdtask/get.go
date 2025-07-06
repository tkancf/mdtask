package mdtask

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/output"
	"github.com/tkancf/mdtask/internal/task"
)

var getCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get a single task by ID",
	Long:  `Get detailed information about a single task by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}

	// Find the task
	foundTask, err := ctx.Repo.FindByID(taskID)
	if err != nil {
		if outputFormat == "json" {
			// For JSON output, return empty object with error
			fmt.Fprintf(os.Stderr, "{\"error\": \"task not found: %s\"}\n", taskID)
			os.Exit(1)
		}
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Output based on format
	if outputFormat == "json" {
		printer := output.NewJSONPrinter(os.Stdout)
		return printer.PrintTask(foundTask)
	}

	// Text output
	printTaskDetails(foundTask)
	return nil
}

func printTaskDetails(t interface{}) {
	// Type assert to access task fields
	task := t.(*task.Task)
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Title: %s\n", task.Title)
	fmt.Printf("Status: %s\n", task.GetStatus())
	
	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}
	
	fmt.Printf("Created: %s\n", task.Created.Format("2006-01-02 15:04"))
	fmt.Printf("Updated: %s\n", task.Updated.Format("2006-01-02 15:04"))
	
	if d := task.GetDeadline(); d != nil {
		fmt.Printf("Deadline: %s", d.Format("2006-01-02"))
		if d.Before(time.Now()) {
			fmt.Printf(" (overdue)")
		}
		fmt.Println()
	}
	
	if r := task.GetReminder(); r != nil {
		fmt.Printf("Reminder: %s\n", r.Format("2006-01-02 15:04"))
	}
	
	if task.IsArchived() {
		fmt.Printf("Archived: Yes\n")
	}
	
	if len(task.Tags) > 0 {
		fmt.Printf("Tags: %v\n", task.Tags)
	}
	
	if task.Content != "" {
		fmt.Printf("\n--- Content ---\n%s\n", task.Content)
	}
}

