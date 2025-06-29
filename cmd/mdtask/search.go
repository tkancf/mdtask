package mdtask

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks by content or tags",
	Long: `Search for tasks by content, title, description, or tag combinations.

Examples:
  # Search by text
  mdtask search "bug fix"
  
  # Search by tags (AND mode - must have all tags)
  mdtask search --tags "type/bug,priority/high"
  
  # Search by tags (OR mode - must have at least one tag)
  mdtask search --tags "type/bug,type/feature" --or
  
  # Exclude specific tags
  mdtask search --tags "type/bug" --exclude "status/done"
  
  # Complex search: text + tags
  mdtask search "login" --tags "type/bug" --exclude "archived"`,
	RunE: runSearch,
}

var (
	searchTags     []string
	excludeTags    []string
	searchOrMode   bool
	searchArchived bool
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringSliceVarP(&searchTags, "tags", "t", []string{}, "Tags to include (comma-separated)")
	searchCmd.Flags().StringSliceVarP(&excludeTags, "exclude", "e", []string{}, "Tags to exclude (comma-separated)")
	searchCmd.Flags().BoolVarP(&searchOrMode, "or", "o", false, "Use OR logic for tags (default is AND)")
	searchCmd.Flags().BoolVarP(&searchArchived, "archived", "a", false, "Include archived tasks")
}

func runSearch(cmd *cobra.Command, args []string) error {
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	var tasks []*task.Task
	var err error

	// If we have tag filters, use tag search
	if len(searchTags) > 0 || len(excludeTags) > 0 {
		// Add archived tag to exclude list if not including archived
		if !searchArchived {
			excludeTags = append(excludeTags, "mdtask/archived")
		}
		
		tasks, err = repo.SearchByTags(searchTags, excludeTags, searchOrMode)
		if err != nil {
			return fmt.Errorf("failed to search by tags: %w", err)
		}
		
		// If we also have a text query, filter results
		if len(args) > 0 {
			query := strings.ToLower(strings.Join(args, " "))
			var filtered []*task.Task
			for _, t := range tasks {
				if strings.Contains(strings.ToLower(t.Title), query) ||
					strings.Contains(strings.ToLower(t.Description), query) ||
					strings.Contains(strings.ToLower(t.Content), query) {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
		}
	} else if len(args) > 0 {
		// Text search only
		query := strings.Join(args, " ")
		tasks, err = repo.Search(query)
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}
		
		// Filter out archived unless requested
		if !searchArchived {
			var filtered []*task.Task
			for _, t := range tasks {
				if !t.IsArchived() {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
		}
	} else {
		// No search criteria
		fmt.Println("Please provide search text or tag filters")
		return nil
	}

	// Display results
	if len(tasks) == 0 {
		fmt.Println("No tasks found matching your criteria")
		return nil
	}

	fmt.Printf("Found %d task(s):\n\n", len(tasks))
	
	// Display search criteria
	if len(args) > 0 {
		fmt.Printf("Text: \"%s\"\n", strings.Join(args, " "))
	}
	if len(searchTags) > 0 {
		mode := "AND"
		if searchOrMode {
			mode = "OR"
		}
		fmt.Printf("Tags (%s): %s\n", mode, strings.Join(searchTags, ", "))
	}
	if len(excludeTags) > 0 {
		// Remove auto-added archived tag from display
		displayExclude := []string{}
		for _, tag := range excludeTags {
			if tag != "mdtask/archived" || searchArchived {
				displayExclude = append(displayExclude, tag)
			}
		}
		if len(displayExclude) > 0 {
			fmt.Printf("Exclude: %s\n", strings.Join(displayExclude, ", "))
		}
	}
	fmt.Println(strings.Repeat("-", 80))

	// Display tasks
	for _, t := range tasks {
		status := string(t.GetStatus())
		deadline := ""
		if d := t.GetDeadline(); d != nil {
			deadline = d.Format("2006-01-02")
			if d.Before(time.Now()) {
				deadline += " (overdue)"
			}
		}
		
		fmt.Printf("[%s] %s\n", status, t.Title)
		if t.Description != "" {
			fmt.Printf("     %s\n", t.Description)
		}
		if deadline != "" {
			fmt.Printf("     Due: %s\n", deadline)
		}
		if len(t.Tags) > 0 {
			// Filter out mdtask system tags for display
			var displayTags []string
			for _, tag := range t.Tags {
				if tag != "mdtask" && !strings.HasPrefix(tag, "mdtask/") {
					displayTags = append(displayTags, tag)
				}
			}
			if len(displayTags) > 0 {
				fmt.Printf("     Tags: %s\n", strings.Join(displayTags, ", "))
			}
		}
		fmt.Println()
	}

	return nil
}