package mdtask

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mdtask",
		Short: "A task management tool using Markdown files",
		Long: `mdtask is a task management tool that treats Markdown files as task tickets.
It provides a CLI interface for managing tasks with YAML frontmatter metadata.`,
	}
	
	// Global output format flag
	outputFormat string
	
	// Version information
	versionInfo = struct {
		Version   string
		Commit    string
		BuildTime string
	}{
		Version:   "dev",
		Commit:    "unknown",
		BuildTime: "unknown",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSlice("paths", []string{"."}, "Paths to search for task files")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text, json)")
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(version, commit, buildTime string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.BuildTime = buildTime
}