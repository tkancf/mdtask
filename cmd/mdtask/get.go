package mdtask

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/repository"
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

	// Load configuration
	cfg, err := config.LoadFromDefaultLocation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	paths, _ := cmd.Flags().GetStringSlice("paths")
	if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
		paths = cfg.Paths
	}
	repo := repository.NewTaskRepository(paths)

	// Find the task
	foundTask, err := repo.FindByID(taskID)
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
		return printTaskJSON(foundTask)
	}

	// Text output
	printTaskDetails(foundTask)
	return nil
}

func printTaskDetails(t *task.Task) {
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Title: %s\n", t.Title)
	fmt.Printf("Status: %s\n", t.GetStatus())
	
	if t.Description != "" {
		fmt.Printf("Description: %s\n", t.Description)
	}
	
	fmt.Printf("Created: %s\n", t.Created.Format("2006-01-02 15:04"))
	fmt.Printf("Updated: %s\n", t.Updated.Format("2006-01-02 15:04"))
	
	if d := t.GetDeadline(); d != nil {
		fmt.Printf("Deadline: %s", d.Format("2006-01-02"))
		if d.Before(time.Now()) {
			fmt.Printf(" (overdue)")
		}
		fmt.Println()
	}
	
	if r := t.GetReminder(); r != nil {
		fmt.Printf("Reminder: %s\n", r.Format("2006-01-02 15:04"))
	}
	
	if t.IsArchived() {
		fmt.Printf("Archived: Yes\n")
	}
	
	if len(t.Tags) > 0 {
		fmt.Printf("Tags: %v\n", t.Tags)
	}
	
	if t.Content != "" {
		fmt.Printf("\n--- Content ---\n%s\n", t.Content)
	}
}

// GetResultJSON represents get result in JSON format
type GetResultJSON struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Reminder    *time.Time `json:"reminder,omitempty"`
	IsArchived  bool       `json:"is_archived"`
	Content     string     `json:"content"`
}

func printTaskJSON(t *task.Task) error {
	result := GetResultJSON{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.GetStatus()),
		Tags:        t.Tags,
		Created:     t.Created,
		Updated:     t.Updated,
		Deadline:    t.GetDeadline(),
		Reminder:    t.GetReminder(),
		IsArchived:  t.IsArchived(),
		Content:     t.Content,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}