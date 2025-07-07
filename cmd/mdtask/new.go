package mdtask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/output"
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
	newCmd.Flags().StringVar(&newParent, "parent", "", "Parent task ID for creating subtask")
}

func runNew(cmd *cobra.Command, args []string) error {
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	// Check if title flag was explicitly provided (even if empty)
	titleFlagProvided := cmd.Flags().Changed("title")
	
	if !titleFlagProvided {
		fmt.Print("Title: ")
		title, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		newTitle = strings.TrimSpace(title)
	}
	// If title flag was provided (including empty string), use as-is
	
	// Apply title prefix from config
	if ctx.Config.Task.TitlePrefix != "" {
		newTitle = ctx.Config.Task.TitlePrefix + newTitle
	}
	
	// Validate title
	if err := task.ValidateTitle(newTitle); err != nil {
		return fmt.Errorf("invalid title: %w", err)
	}

	// Check if description flag was explicitly provided (even if empty)
	descriptionFlagProvided := cmd.Flags().Changed("description")
	
	if !descriptionFlagProvided {
		// Use description template if available
		if ctx.Config.Task.DescriptionTemplate != "" {
			newDescription = ctx.Config.Task.DescriptionTemplate
		} else {
			// Default to empty string instead of prompting
			newDescription = ""
		}
	}
	// If description flag was provided (including empty string), use as-is
	
	// Validate description
	if err := task.ValidateDescription(newDescription); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	var content string
	// Check if content flag was explicitly provided (even if empty)
	contentFlagProvided := cmd.Flags().Changed("content")
	
	if contentFlagProvided {
		// Use content from flag (including empty string)
		content = newContent
	} else if ctx.Config.Task.ContentTemplate != "" {
		// Use content template if available
		content = ctx.Config.Task.ContentTemplate
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
	allTags = append(allTags, ctx.Config.Task.DefaultTags...)
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

	// Add title as first line of content
	titleLine := fmt.Sprintf("# %s", newTitle)
	if content != "" {
		content = fmt.Sprintf("%s\n\n%s", titleLine, content)
	} else {
		content = titleLine
	}
	t.Content = content

	// Use config default status if not specified
	statusStr := newStatus
	if statusStr == "" {
		statusStr = ctx.Config.Task.DefaultStatus
	}
	if statusStr == "" {
		statusStr = "TODO"
	}
	statusToSet, err := cli.ValidateStatus(statusStr)
	if err != nil {
		return err
	}
	t.SetStatus(statusToSet)

	if deadline, err := cli.ParseDeadline(newDeadline); err != nil {
		return err
	} else if deadline != nil {
		t.SetDeadline(*deadline)
	}

	if reminder, err := cli.ParseReminder(newReminder); err != nil {
		return err
	} else if reminder != nil {
		t.SetReminder(*reminder)
	}

	// Handle parent task relationship
	if newParent != "" {
		// Normalize parent task ID
		normalizedParent, err := cli.NormalizeTaskID(newParent)
		if err != nil {
			return err
		}
		newParent = normalizedParent
		
		// Validate parent task exists
		parentTask, err := ctx.Repo.FindByID(newParent)
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

	filePath, err := ctx.Repo.Create(t)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	if outputFormat == "json" {
		printer := output.NewJSONPrinter(os.Stdout)
		return printer.PrintTaskWithPath(t, filePath)
	}

	fmt.Printf("\nTask created successfully!\n")
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("File: %s\n", filePath)

	return nil
}