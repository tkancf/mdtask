package mdtask

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/repository"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks",
	Long:  `Search tasks by title, description, content, or tags.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")
	
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	tasks, err := repo.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Printf("No tasks found matching '%s'.\n", query)
		return nil
	}

	fmt.Printf("Found %d task(s) matching '%s':\n\n", len(tasks), query)
	printTasks(tasks)
	return nil
}