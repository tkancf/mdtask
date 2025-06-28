package mdtask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/repository"
	"github.com/tkan/mdtask/internal/task"
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
	newTags        []string
	newStatus      string
	newDeadline    string
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&newTitle, "title", "t", "", "Task title")
	newCmd.Flags().StringVarP(&newDescription, "description", "d", "", "Task description")
	newCmd.Flags().StringSliceVar(&newTags, "tags", []string{}, "Additional tags (comma-separated)")
	newCmd.Flags().StringVarP(&newStatus, "status", "s", "TODO", "Initial status (TODO, WIP, WAIT, SCHE, DONE)")
	newCmd.Flags().StringVar(&newDeadline, "deadline", "", "Deadline (YYYY-MM-DD)")
}

func runNew(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	if newTitle == "" {
		fmt.Print("Title: ")
		title, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		newTitle = strings.TrimSpace(title)
	}

	if newDescription == "" {
		fmt.Print("Description: ")
		desc, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		newDescription = strings.TrimSpace(desc)
	}

	fmt.Println("\nContent (press Ctrl+D to finish):")
	var contentBuilder strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		contentBuilder.WriteString(scanner.Text())
		contentBuilder.WriteString("\n")
	}

	now := time.Now()
	t := &task.Task{
		Title:       newTitle,
		Description: newDescription,
		Created:     now,
		Updated:     now,
		Content:     strings.TrimSpace(contentBuilder.String()),
		Tags:        append([]string{"mdtask"}, newTags...),
		Aliases:     []string{},
	}

	status := task.Status(newStatus)
	t.SetStatus(status)

	if newDeadline != "" {
		deadline, err := time.Parse("2006-01-02", newDeadline)
		if err != nil {
			return fmt.Errorf("invalid deadline format: %w", err)
		}
		t.SetDeadline(deadline)
	}

	paths, _ := cmd.Flags().GetStringSlice("paths")
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