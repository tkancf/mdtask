package mdtask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new task",
	Long:  `Create a new task interactively or with provided flags.`,
	RunE:  runNew,
}

var (
	newTitle       string
	newDescription string
	newContent     string
	newTags        []string
	newStatus      string
	newDeadline    string
	newReminder    string
	newParent      string
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&newTitle, "title", "t", "", "Task title")
	newCmd.Flags().StringVarP(&newDescription, "description", "d", "", "Task description")
	newCmd.Flags().StringVarP(&newContent, "content", "c", "", "Task content")
	newCmd.Flags().StringSliceVar(&newTags, "tags", []string{}, "Additional tags (comma-separated)")
	newCmd.Flags().StringVarP(&newStatus, "status", "s", "", "Initial status (TODO, WIP, WAIT, SCHE, DONE)")
	newCmd.Flags().StringVar(&newDeadline, "deadline", "", "Deadline (YYYY-MM-DD)")
	newCmd.Flags().StringVar(&newReminder, "reminder", "", "Reminder (YYYY-MM-DD HH:MM or YYYY-MM-DD)")
}

func runNew(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromDefaultLocation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	if newTitle == "" {
		fmt.Print("Title: ")
		title, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		newTitle = strings.TrimSpace(title)
	}
	
	// Apply title prefix from config
	if cfg.Task.TitlePrefix != "" {
		newTitle = cfg.Task.TitlePrefix + newTitle
	}
	
	// Validate title
	if err := task.ValidateTitle(newTitle); err != nil {
		return fmt.Errorf("invalid title: %w", err)
	}

	if newDescription == "" {
		// Use description template if available
		if cfg.Task.DescriptionTemplate != "" {
			newDescription = cfg.Task.DescriptionTemplate
		} else {
			fmt.Print("Description: ")
			desc, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			newDescription = strings.TrimSpace(desc)
		}
	}
	
	// Validate description
	if err := task.ValidateDescription(newDescription); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	var content string
	if newContent != "" {
		// Use content from flag
		content = newContent
	} else if cfg.Task.ContentTemplate != "" {
		// Use content template if available
		content = cfg.Task.ContentTemplate
		fmt.Println("\nUsing content template from configuration.")
	} else {
		fmt.Println("\nContent (press Ctrl+D to finish):")
		var contentBuilder strings.Builder
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			contentBuilder.WriteString(scanner.Text())
			contentBuilder.WriteString("\n")
		}
		content = strings.TrimSpace(contentBuilder.String())
	}

	now := time.Now()
	
	// Merge default tags from config with provided tags
	allTags := []string{"mdtask"}
	allTags = append(allTags, cfg.Task.DefaultTags...)
	allTags = append(allTags, newTags...)
	
	t := &task.Task{
		Title:       newTitle,
		Description: newDescription,
		Created:     now,
		Updated:     now,
		Content:     content,
		Tags:        allTags,
		Aliases:     []string{},
	}

	// Use config default status if not specified
	statusStr := newStatus
	if statusStr == "" {
		statusStr = cfg.Task.DefaultStatus
	}
	if statusStr == "" {
		statusStr = "TODO"
	}
	t.SetStatus(task.Status(statusStr))

	if newDeadline != "" {
		deadline, err := time.Parse("2006-01-02", newDeadline)
		if err != nil {
			return fmt.Errorf("invalid deadline format: %w", err)
		}
		t.SetDeadline(deadline)
	}

	if newReminder != "" {
		// Try parsing with time first
		reminder, err := time.Parse("2006-01-02 15:04", newReminder)
		if err != nil {
			// Try parsing date only
			reminder, err = time.Parse("2006-01-02", newReminder)
			if err != nil {
				return fmt.Errorf("invalid reminder format (use YYYY-MM-DD HH:MM or YYYY-MM-DD): %w", err)
			}
			// Set default time to 9:00 AM for date-only reminders
			reminder = time.Date(reminder.Year(), reminder.Month(), reminder.Day(), 9, 0, 0, 0, reminder.Location())
		}
		t.SetReminder(reminder)
	}

	// Handle parent task relationship
	if newParent != "" {
		// Validate parent task ID format
		if !strings.HasPrefix(newParent, "task/") {
			// Try to add the prefix if it's missing
			if _, err := time.Parse("20060102150405", newParent); err == nil {
				newParent = "task/" + newParent
			} else {
				return fmt.Errorf("invalid parent task ID format: %s", newParent)
			}
		}
		
		// Validate parent task exists
		paths, _ := cmd.Flags().GetStringSlice("paths")
		if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
			paths = cfg.Paths
		}
		tempRepo := repository.NewTaskRepository(paths)
		parentTask, err := tempRepo.FindByID(newParent)
		if err != nil {
			return fmt.Errorf("parent task not found: %s", newParent)
		}
		
		// Set parent ID
		t.SetParentID(newParent)
		
		// Optionally inherit some properties from parent
		if newStatus == "" && statusStr == "TODO" {
			// Inherit parent's status if not specified
			t.SetStatus(parentTask.GetStatus())
		}
		
		fmt.Printf("Creating subtask of: %s\n", parentTask.Title)
	}

	paths, _ := cmd.Flags().GetStringSlice("paths")
	if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
		paths = cfg.Paths
	}
	repo := repository.NewTaskRepository(paths)

	filePath, err := repo.Create(t)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("\nTask created successfully!\n")
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("File: %s\n", filePath)

	return nil
}