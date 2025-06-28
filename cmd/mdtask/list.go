package mdtask

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/config"
	"github.com/tkan/mdtask/internal/repository"
	"github.com/tkan/mdtask/internal/task"
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
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (TODO, WIP, WAIT, SCHE, DONE)")
	listCmd.Flags().BoolVarP(&listArchived, "archived", "a", false, "Show only archived tasks")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all tasks including archived")
}

func runList(cmd *cobra.Command, args []string) error {
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

	var tasks []*task.Task

	if listStatus != "" {
		status := task.Status(listStatus)
		tasks, err = repo.FindByStatus(status)
	} else if listArchived {
		allTasks, err := repo.FindAll()
		if err != nil {
			return err
		}
		for _, t := range allTasks {
			if t.IsArchived() {
				tasks = append(tasks, t)
			}
		}
	} else if listAll {
		tasks, err = repo.FindAll()
	} else {
		tasks, err = repo.FindActive()
	}

	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
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

	fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDEADLINE\tARCHIVED")
	fmt.Fprintln(w, strings.Repeat("-", 80))

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

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			t.GetStatus(),
			truncate(t.Title, 40),
			deadline,
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