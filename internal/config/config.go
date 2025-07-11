package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the mdtask configuration
type Config struct {
	// Paths to search for task files
	Paths []string `toml:"paths"`
	
	// Task creation settings
	Task TaskConfig `toml:"task"`
	
	// Web server settings
	Web WebConfig `toml:"web"`
	
	// MCP server settings
	MCP MCPConfig `toml:"mcp"`
	
	// Editor settings
	Editor EditorConfig `toml:"editor"`
}

// TaskConfig contains task-related configuration
type TaskConfig struct {
	// Prefix to add to task titles
	TitlePrefix string `toml:"title_prefix"`
	
	// Default status for new tasks
	DefaultStatus string `toml:"default_status"`
	
	// Template for new task content
	ContentTemplate string `toml:"content_template"`
	
	// Template for new task description
	DescriptionTemplate string `toml:"description_template"`
	
	// Default tags to add to new tasks
	DefaultTags []string `toml:"default_tags"`
}

// WebConfig contains web server configuration
type WebConfig struct {
	// Default port for web server
	Port int `toml:"port"`
	
	// Whether to open browser automatically
	OpenBrowser bool `toml:"open_browser"`
}

// MCPConfig contains MCP server configuration
type MCPConfig struct {
	// Whether MCP server is enabled
	Enabled bool `toml:"enabled"`
	
	// Additional allowed paths for MCP server
	AllowedPaths []string `toml:"allowed_paths"`
}

// EditorConfig contains editor configuration
type EditorConfig struct {
	// Command to launch editor (e.g., "vim", "code", "emacs")
	// If empty, uses $EDITOR environment variable
	Command string `toml:"command"`
	
	// Arguments to pass to editor command
	Args []string `toml:"args"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Paths: []string{"."},
		Task: TaskConfig{
			TitlePrefix:   "",
			DefaultStatus: "TODO",
		},
		Web: WebConfig{
			Port:        7000,
			OpenBrowser: true,
		},
		MCP: MCPConfig{
			Enabled:      true,
			AllowedPaths: []string{},
		},
		Editor: EditorConfig{
			Command: "", // Will use $EDITOR by default
			Args:    []string{},
		},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	config := DefaultConfig()
	
	// Try to read the file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse TOML
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return config, nil
}

// FindConfigFile searches for config file in standard locations
func FindConfigFile() string {
	// Check current directory
	if _, err := os.Stat(".mdtask.toml"); err == nil {
		return ".mdtask.toml"
	}
	
	if _, err := os.Stat("mdtask.toml"); err == nil {
		return "mdtask.toml"
	}
	
	// Check home directory
	if home, err := os.UserHomeDir(); err == nil {
		configPath := filepath.Join(home, ".config", "mdtask", "config.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		
		configPath = filepath.Join(home, ".mdtask.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	
	return ""
}

// LoadFromDefaultLocation loads config from default locations
func LoadFromDefaultLocation() (*Config, error) {
	configFile := FindConfigFile()
	if configFile == "" {
		return DefaultConfig(), nil
	}
	
	return Load(configFile)
}

// GetEditor returns the editor command and arguments
func (c *Config) GetEditor() (string, []string) {
	if c.Editor.Command != "" {
		return c.Editor.Command, c.Editor.Args
	}
	
	// Fall back to $EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, []string{}
	}
	
	// Default to vim if nothing is configured
	return "vim", []string{}
}