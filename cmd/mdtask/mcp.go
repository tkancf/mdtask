package mdtask

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/config"
	mcpserver "github.com/tkan/mdtask/internal/mcp"
	"github.com/tkan/mdtask/internal/repository"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes mdtask functionality to AI assistants.

This server implements the Model Context Protocol, allowing AI tools like
Claude Desktop to interact with your tasks through a standardized interface.

The server provides:
- Tools for creating, updating, and managing tasks
- Resources for accessing task lists and statistics
- Full integration with your mdtask configuration

To use with Claude Desktop, add this to your claude_desktop_config.json:

  "mcpServers": {
    "mdtask": {
      "command": "mdtask",
      "args": ["mcp"],
      "cwd": "/path/to/your/tasks"
    }
  }`,
	Run: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadFromDefaultLocation()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create repository
	repo := repository.NewTaskRepository(cfg.Paths)

	// Create and start MCP server
	server := mcpserver.NewServer(repo, cfg)
	
	// Log to stderr since stdout is used for MCP protocol
	fmt.Fprintln(cmd.ErrOrStderr(), "Starting mdtask MCP server...")
	
	if err := server.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}