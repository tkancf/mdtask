# mdtask

- mdtask is a tool for managing Markdown files as task management tickets

## Markdown File Format

Markdown files used as task management tickets have YAML frontmatter and include the mdtask tag in tags

```yaml
---
id: unique-identifier
aliases:
  - Alternative Title
tags:
  - mdtask
created: YYYY-MM-DD HH:MM
description: Brief description
title: Display Title
updated: YYYY-MM-DD HH:MM
---

# Content Here
```

- unique-identifier = task/YYYYMMDDHHMMSS
- YYYYMMDDHHMMSS is the file creation date and time

### Task Management

Task status and management scope are determined using tags in the YAML frontmatter

- If the file has the mdtask tag
    - It is managed by mdtask
- Task status is managed with `mdtsk/status/***`
    - TODO status: mdtask/status/TODO
    - Work in progress status: mdtask/status/WIP
    - Waiting for response status: mdtask/status/WAIT
    - Scheduled and waiting status: mdtask/status/SCHE
    - Completed status: mdtask/status/DONE
- Task archiving is managed with `mdtask/archived`
    - Archived task: mdtask/archived
- Task deadlines are managed with `mdtask/deadline/YYYY-MM-DD`
    - Task with deadline 2025/06/29: mdtask/deadline/2025-06-29
- Reasons for waiting status (`mdtask/status/WAIT`) are managed with `mdtask/waitfor/****`
    - Task waiting for email reply: `mdtask/waitfor/waiting-for-email-reply`

## mdtask Features

- Manages and creates Markdown files in the above format
- Implemented in Go
- mdtask provides a CLI interface
    - `mdtask list` - List tasks (with --status, --archived, --all options)
    - `mdtask search [query]` - Search tasks
    - `mdtask new` - Create a new task (interactive or with flags)
    - `mdtask edit [task-id]` - Edit a task (launches editor)
    - `mdtask archive [task-id]` - Archive a task
    - `mdtask tui` - Launch terminal UI (interactive task management)
- mdtask provides a web browser interface
    - `mdtask web` - Launch WebUI (default port: 7000, with automatic port switching)
    - Intuitive UI including dashboard, task management, and search functionality
- mdtask provides an MCP (Model Context Protocol) server
    - `mdtask mcp` - Launch MCP server (for AI assistants)
    - Manage tasks from MCP-compatible tools like Claude Desktop
- mdtask configuration
    - Supports TOML configuration files (.mdtask.toml, mdtask.toml, ~/.config/mdtask/config.toml, ~/.mdtask.toml)
    - Configurable options:
        - `paths` - Specify managed directories
        - `task.title_prefix` - Prefix automatically added to task titles
        - `task.default_status` - Default status for new tasks
        - `web.port` - Default port number for WebUI
        - `web.open_browser` - Auto-launch browser when starting WebUI
        - `mcp.enabled` - Enable/disable MCP server
        - `mcp.allowed_paths` - Additional paths accessible by MCP server
        - `editor.command` - Editor command for task editing (uses $EDITOR if not set)
        - `editor.args` - Additional arguments to pass to the editor

## Installation

### Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher (for WebUI style and JavaScript generation)

### Download Pre-built Binaries

Pre-built binaries are available for macOS, Linux, and Windows from the [releases page](https://github.com/tkancf/mdtask/releases).

Download the appropriate binary for your platform:
- `mdtask-darwin-amd64` - macOS (Intel)
- `mdtask-darwin-arm64` - macOS (Apple Silicon)
- `mdtask-linux-amd64` - Linux (x86_64)
- `mdtask-linux-arm64` - Linux (ARM64)
- `mdtask-windows-amd64.exe` - Windows (x86_64)

After downloading, make the binary executable (macOS/Linux):
```bash
chmod +x mdtask-*
sudo mv mdtask-* /usr/local/bin/mdtask
```

### Build from Source

```bash
git clone https://github.com/tkancf/mdtask.git
cd mdtask

# Install dependencies and build
make

# Or run individually
npm install
npm run build
go build -o mdtask
```

### Makefile Targets

- `make` - Install dependencies, generate CSS/JavaScript, build binary
- `make build` - Build binary (including CSS/JavaScript generation)
- `make css` - Build CSS only
- `make js` - Build JavaScript only (TypeScript compilation)
- `make watch` - Watch CSS changes (for development)
- `make test` - Run tests
- `make release` - Release build for all platforms
- `make clean` - Clean build artifacts
- `make install` - Local installation (/usr/local/bin)

### Development Mode

During development, you can watch CSS/JavaScript changes:

```bash
# Watch CSS changes
npm run watch-css

# In another terminal, watch TypeScript/JavaScript changes
npm run dev-js

# In yet another terminal, run the Go application
go run main.go web
```

### Tech Stack

- **Backend**: Go 1.19+
  - Cobra (CLI framework)
  - Chi (HTTP router)
  - mark3labs/mcp-go (MCP implementation)
- **Frontend**: 
  - TypeScript (type-safe JavaScript)
  - Vite (fast build tool)
  - Tailwind CSS (utility-first CSS)
- **Build Tools**: 
  - Make (build automation)
  - npm (package management)

## MCP (Model Context Protocol) Server

mdtask includes a built-in MCP server, allowing you to manage tasks from MCP-compatible AI assistants like Claude Desktop.

### MCP Configuration

To use with Claude Desktop, add the following to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mdtask": {
      "command": "/path/to/mdtask",
      "args": ["mcp"],
      "cwd": "/path/to/your/tasks"
    }
  }
}
```

### Available MCP Tools

- `list_tasks` - List tasks (with status filter and archive display support)
- `create_task` - Create a new task
- `update_task` - Update task (title, description, status, tags)
- `search_tasks` - Search tasks
- `archive_task` - Archive a task
- `get_task` - Get details of a specific task
- `get_statistics` - Get task statistics

### Available MCP Resources

- `tasks` - Markdown-formatted list of active tasks
- `statistics` - Task statistics (JSON format)

## Neovim Plugin

mdtask includes a plugin that allows you to manage tasks directly from Neovim.

### Installation

The plugin is located in the `nvim-mdtask` subdirectory.

**For lazy.nvim:**
```lua
{
  dir = '~/path/to/mdtask/nvim-mdtask',  -- Specify the path to your mdtask repository
  name = 'nvim-mdtask',
  dependencies = {
    'nvim-telescope/telescope.nvim', -- optional
  },
  config = function()
    require('mdtask').setup()
  end,
}
```

### Main Features

- `:MdTask` - Display task list
- `:MdTask new` - Create new task
- `:MdTask search <query>` - Search tasks
- `:MdTask status <status>` - Display by status

For details, see [nvim-mdtask/README.md](nvim-mdtask/README.md).

## Development

### Architecture

mdtask consists of the following layers:

- **CLI Command Layer** (`cmd/mdtask/`): User interface
- **Service Layer** (`internal/service/`): Business logic
- **Repository Layer** (`internal/repository/`): Data access
- **Common Utilities** (`internal/cli/`, `internal/output/`): Cross-cutting concerns

### Build and Test

```bash
# Build
go build -o mdtask

# Run tests
go test ./...

# Run tests and lint
./test.sh
```

### Code Structure

```
mdtask/
├── cmd/mdtask/          # CLI commands
│   ├── root.go         # Root command
│   ├── new.go          # Task creation
│   ├── list.go         # Task listing
│   └── ...
├── internal/           # Internal packages
│   ├── cli/           # CLI common utilities
│   ├── service/       # Business logic layer
│   ├── repository/    # Data access layer
│   ├── task/          # Task model
│   └── config/        # Configuration management
├── pkg/               # Public packages
│   └── markdown/      # Markdown parser
└── nvim-mdtask/       # Neovim plugin
```

### Contributing

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

### Creating a Release

Releases are automatically built and published when a new tag is pushed:

```bash
# Tag a new version
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow will:
- Build binaries for all supported platforms
- Generate checksums
- Create a GitHub release with auto-generated release notes
- Upload all binaries and checksums as release assets
