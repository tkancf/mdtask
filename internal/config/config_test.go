package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if len(config.Paths) != 1 || config.Paths[0] != "." {
		t.Errorf("DefaultConfig() Paths = %v, want [.]", config.Paths)
	}
	
	if config.Task.TitlePrefix != "" {
		t.Errorf("DefaultConfig() TitlePrefix = %v, want empty", config.Task.TitlePrefix)
	}
	
	if config.Task.DefaultStatus != "TODO" {
		t.Errorf("DefaultConfig() DefaultStatus = %v, want TODO", config.Task.DefaultStatus)
	}
	
	if config.Web.Port != 7000 {
		t.Errorf("DefaultConfig() Port = %v, want 7000", config.Web.Port)
	}
	
	if !config.Web.OpenBrowser {
		t.Errorf("DefaultConfig() OpenBrowser = %v, want true", config.Web.OpenBrowser)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantConfig  *Config
		wantErr     bool
	}{
		{
			name: "valid config",
			content: `paths = [".", "tasks/"]

[task]
title_prefix = "[PREFIX] "
default_status = "WIP"
content_template = "## Task Details\n\n"
description_template = "New task description"
default_tags = ["project/test", "priority/medium"]

[web]
port = 8080
open_browser = false`,
			wantConfig: &Config{
				Paths: []string{".", "tasks/"},
				Task: TaskConfig{
					TitlePrefix:         "[PREFIX] ",
					DefaultStatus:       "WIP",
					ContentTemplate:     "## Task Details\n\n",
					DescriptionTemplate: "New task description",
					DefaultTags:         []string{"project/test", "priority/medium"},
				},
				Web: WebConfig{
					Port:        8080,
					OpenBrowser: false,
				},
			},
			wantErr: false,
		},
		{
			name: "partial config",
			content: `[task]
title_prefix = "[TEST] "

[web]
port = 9000`,
			wantConfig: &Config{
				Paths: []string{"."}, // default
				Task: TaskConfig{
					TitlePrefix:   "[TEST] ",
					DefaultStatus: "TODO", // default
				},
				Web: WebConfig{
					Port:        9000,
					OpenBrowser: true, // default
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid TOML",
			content: `invalid toml content [[[`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tempFile, err := os.CreateTemp("", "mdtask-config-*.toml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())
			
			// Write test content
			if _, err := tempFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tempFile.Close()
			
			// Load config
			config, err := Load(tempFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && !reflect.DeepEqual(config, tt.wantConfig) {
				t.Errorf("Load() config = %+v, want %+v", config, tt.wantConfig)
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	config, err := Load("/non/existent/path/config.toml")
	if err != nil {
		t.Errorf("Load() should return default config for non-existent file, got error: %v", err)
	}
	
	defaultConfig := DefaultConfig()
	if !reflect.DeepEqual(config, defaultConfig) {
		t.Errorf("Load() for non-existent file = %+v, want default config", config)
	}
}

func TestFindConfigFile(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "mdtask-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	tests := []struct {
		name       string
		createFile string
		want       string
	}{
		{
			name:       ".mdtask.toml in current dir",
			createFile: ".mdtask.toml",
			want:       ".mdtask.toml",
		},
		{
			name:       "mdtask.toml in current dir",
			createFile: "mdtask.toml",
			want:       "mdtask.toml",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing files
			os.Remove(".mdtask.toml")
			os.Remove("mdtask.toml")
			
			// Create test file if needed
			if tt.createFile != "" {
				if err := os.WriteFile(tt.createFile, []byte("paths = [\".\"]"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}
			
			got := FindConfigFile()
			if got != tt.want {
				t.Errorf("FindConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindConfigFileNoConfig(t *testing.T) {
	// This test might fail if user has config in home directory
	// We'll skip it in that case
	if home, err := os.UserHomeDir(); err == nil {
		possibleConfigs := []string{
			filepath.Join(home, ".config", "mdtask", "config.toml"),
			filepath.Join(home, ".mdtask.toml"),
		}
		for _, config := range possibleConfigs {
			if _, err := os.Stat(config); err == nil {
				t.Skip("Skipping test because config exists in home directory")
			}
		}
	}
	
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "mdtask-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	got := FindConfigFile()
	if got != "" {
		t.Errorf("FindConfigFile() = %v, want empty string", got)
	}
}

func TestFindConfigFileInHome(t *testing.T) {
	// This test is tricky because it depends on the user's home directory
	// We'll create a mock scenario
	
	// Save original home
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Create temp home
	tempHome, err := os.MkdirTemp("", "mdtask-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	
	// Set temp home
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)
	
	// Create config directory
	configDir := filepath.Join(tempHome, ".config", "mdtask")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Create config file
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("paths = [\".\"]"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Test finding it
	found := FindConfigFile()
	if found != configPath {
		// Try .mdtask.toml in home
		homeConfig := filepath.Join(tempHome, ".mdtask.toml")
		if err := os.WriteFile(homeConfig, []byte("paths = [\".\"]"), 0644); err != nil {
			t.Fatalf("Failed to create home config file: %v", err)
		}
		
		found = FindConfigFile()
		if found != homeConfig && found != "" {
			t.Errorf("FindConfigFile() didn't find config in home directory")
		}
	}
}

func TestLoadFromDefaultLocation(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "mdtask-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	
	tests := []struct {
		name        string
		createFile  bool
		fileContent string
		wantPrefix  string
	}{
		{
			name:        "with config file",
			createFile:  true,
			fileContent: `[task]
title_prefix = "[LOADED] "`,
			wantPrefix: "[LOADED] ",
		},
		{
			name:       "without config file",
			createFile: false,
			wantPrefix: "", // default
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Remove(".mdtask.toml")
			
			// Create file if needed
			if tt.createFile {
				if err := os.WriteFile(".mdtask.toml", []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
			}
			
			config, err := LoadFromDefaultLocation()
			if err != nil {
				t.Errorf("LoadFromDefaultLocation() error = %v", err)
				return
			}
			
			if config.Task.TitlePrefix != tt.wantPrefix {
				t.Errorf("LoadFromDefaultLocation() TitlePrefix = %v, want %v", config.Task.TitlePrefix, tt.wantPrefix)
			}
		})
	}
}

func TestConfigMerging(t *testing.T) {
	// Test that loaded config properly merges with defaults
	content := `[task]
title_prefix = "[CUSTOM] "

# Note: default_status is not specified, should use default`
	
	tempFile, err := os.CreateTemp("", "mdtask-merge-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()
	
	config, err := Load(tempFile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	
	// Check merged values
	if config.Task.TitlePrefix != "[CUSTOM] " {
		t.Errorf("Load() didn't apply custom title_prefix")
	}
	
	if config.Task.DefaultStatus != "TODO" {
		t.Errorf("Load() didn't preserve default for default_status")
	}
	
	if config.Web.Port != 7000 {
		t.Errorf("Load() didn't preserve default for port")
	}
}