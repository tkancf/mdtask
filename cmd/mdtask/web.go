package mdtask

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkan/mdtask/internal/config"
	"github.com/tkan/mdtask/internal/constants"
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
	webCmd.Flags().StringVarP(&webPort, "port", "p", strconv.Itoa(constants.DefaultWebPort), "Port to run the web server on")
	webCmd.Flags().BoolVar(&webOpen, "open", constants.DefaultOpenBrowser, "Open browser automatically")
}

func runWeb(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromDefaultLocation()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use config values if not specified by flags
	port := webPort
	if cmd.Flags().Changed("port") == false && cfg.Web.Port > 0 {
		port = strconv.Itoa(cfg.Web.Port)
	}
	
	openBrowser := webOpen
	if cmd.Flags().Changed("open") == false {
		openBrowser = cfg.Web.OpenBrowser
	}

	paths, _ := cmd.Flags().GetStringSlice("paths")
	if len(paths) == 1 && paths[0] == "." && len(cfg.Paths) > 0 {
		paths = cfg.Paths
	}
	
	repo := repository.NewTaskRepository(paths)

	server, err := web.NewServer(repo, cfg, port)
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
	if openBrowser {
		url := fmt.Sprintf("http://localhost:%s", server.GetPort())
		fmt.Printf("Opening browser at %s...\n", url)
		if err := openBrowserURL(url); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
		}
	}

	// Wait for server error
	return <-errCh
}

func openBrowserURL(url string) error {
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