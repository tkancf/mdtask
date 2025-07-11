package mdtask

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/output"
	"github.com/tkancf/mdtask/internal/service"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [task-id]",
	Short: "Archive a task",
	Long:  `Archive a task by adding the mdtask/archived tag.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}
	
	taskID, err := cli.NormalizeTaskID(args[0])
	if err != nil {
		return err
	}
	
	// Use service layer for business logic
	taskService := service.NewTaskService(ctx.Repo, ctx.Config)
	t, err := taskService.ArchiveTask(taskID)
	if err != nil {
		return err
	}
	
	if outputFormat == "json" {
		printer := output.NewJSONPrinter(os.Stdout)
		return printer.PrintTask(t)
	}
	
	fmt.Printf("Task %s archived successfully.\n", taskID)
	return nil
}