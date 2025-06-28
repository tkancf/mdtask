package mdtask

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/repository"
	"github.com/tkan/mdtask/internal/web"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface",
	Long:  `Start the web interface for managing tasks through a browser.`,
	RunE:  runWeb,
}

var (
	webPort string
	webOpen bool
)

func init() {
	rootCmd.AddCommand(webCmd)
	webCmd.Flags().StringVarP(&webPort, "port", "p", "7000", "Port to run the web server on")
	webCmd.Flags().BoolVar(&webOpen, "open", true, "Open browser automatically")
}

func runWeb(cmd *cobra.Command, args []string) error {
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	server, err := web.NewServer(repo, webPort)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	// Start server in a goroutine to handle browser opening after port is determined
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Wait a bit for server to bind to port
	time.Sleep(100 * time.Millisecond)

	// Open browser if requested
	if webOpen {
		url := fmt.Sprintf("http://localhost:%s", server.GetPort())
		fmt.Printf("Opening browser at %s...\n", url)
		if err := openBrowser(url); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
		}
	}

	// Wait for server error
	return <-errCh
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}