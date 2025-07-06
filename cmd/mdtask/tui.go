package mdtask

import (
	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI mode",
	Long:  `Launch an interactive Terminal User Interface for managing tasks`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}

	app := tui.NewApp(ctx.Repo, ctx.Config)
	return app.Run()
}