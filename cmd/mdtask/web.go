package mdtask

import (
	"fmt"
	"os/exec"
	"runtime"

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
	webCmd.Flags().StringVarP(&webPort, "port", "p", "8080", "Port to run the web server on")
	webCmd.Flags().BoolVar(&webOpen, "open", true, "Open browser automatically")
}

func runWeb(cmd *cobra.Command, args []string) error {
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	server, err := web.NewServer(repo, webPort)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	// Open browser if requested
	if webOpen {
		url := fmt.Sprintf("http://localhost:%s", webPort)
		go func() {
			// Wait a bit for server to start
			// In production, we'd check if the server is actually running
			fmt.Printf("Opening browser at %s...\n", url)
			openBrowser(url)
		}()
	}

	return server.Start()
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