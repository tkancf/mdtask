package mdtask

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/output"
	"github.com/tkancf/mdtask/internal/task"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List all tasks in the configured directories.`,
	RunE:  runList,
}

var (
	listStatus   string
	listArchived bool
	listAll      bool
	listParent   string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (TODO, WIP, WAIT, SCHE, DONE)")
	listCmd.Flags().BoolVarP(&listArchived, "archived", "a", false, "Show only archived tasks")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all tasks including archived")
	listCmd.Flags().StringVar(&listParent, "parent", "", "Show only subtasks of the specified parent task ID")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}

	var tasks []*task.Task

	// Handle parent filter
	if listParent != "" {
		// Validate parent task ID format
		if !strings.HasPrefix(listParent, "task/") {
			// Try to add the prefix if it's missing
			if _, err := time.Parse("20060102150405", listParent); err == nil {
				listParent = "task/" + listParent
			} else {
				return fmt.Errorf("invalid parent task ID format: %s", listParent)
			}
		}
		
		// Get all tasks and filter by parent
		allTasks, err := ctx.Repo.FindAll()
		if err != nil {
			return err
		}
		
		for _, t := range allTasks {
			if t.GetParentID() == listParent {
				// Apply additional filters if specified
				if listStatus != "" && string(t.GetStatus()) != listStatus {
					continue
				}
				if listArchived && !t.IsArchived() {
					continue
				}
				if !listAll && !listArchived && t.IsArchived() {
					continue
				}
				tasks = append(tasks, t)
			}
		}
	} else if listStatus != "" {
		status := task.Status(listStatus)
		tasks, err = ctx.Repo.FindByStatus(status)
	} else if listArchived {
		allTasks, err := ctx.Repo.FindAll()
		if err != nil {
			return err
		}
		for _, t := range allTasks {
			if t.IsArchived() {
				tasks = append(tasks, t)
			}
		}
	} else if listAll {
		tasks, err = ctx.Repo.FindAll()
	} else {
		tasks, err = ctx.Repo.FindActive()
	}

	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if outputFormat == "json" {
		printer := output.NewJSONPrinter(os.Stdout)
		if len(tasks) == 0 {
			return printer.PrintEmpty()
		}
		return printer.PrintTasks(tasks)
	}
	
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}
	
	printTasks(tasks)
	return nil
}

func printTasks(tasks []*task.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDEADLINE\tPARENT\tARCHIVED")
	fmt.Fprintln(w, strings.Repeat("-", 90))

	for _, t := range tasks {
		deadline := ""
		if d := t.GetDeadline(); d != nil {
			deadline = d.Format("2006-01-02")
			if d.Before(time.Now()) {
				deadline += " (overdue)"
			}
		}

		archived := ""
		if t.IsArchived() {
			archived = "✓"
		}
		
		parent := ""
		if parentID := t.GetParentID(); parentID != "" {
			// Show only the timestamp part for brevity
			if strings.HasPrefix(parentID, "task/") {
				parent = parentID[5:]
			} else {
				parent = parentID
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			t.GetStatus(),
			truncate(t.Title, 40),
			deadline,
			parent,
			archived,
		)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

