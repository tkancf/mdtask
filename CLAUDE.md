# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

mdtask is a task management tool that uses Markdown files as task tickets. The project is in its initial stages with only documentation committed.

## Technology Stack

- **Language**: Go (planned)
- **Configuration**: TOML format
- **Task Storage**: Markdown files with YAML frontmatter

## Architecture

### Task File Format

Tasks are Markdown files with YAML frontmatter containing:
- **ID**: `task/YYYYMMDDHHMMSS` format (based on creation timestamp)
- **Tags**: Must include `mdtask` to be recognized as a managed task
- **Status Tags**: One of:
  - `mdtask/status/TODO` - To do
  - `mdtask/status/WIP` - Work in progress
  - `mdtask/status/WAIT` - Waiting for response
  - `mdtask/status/SCHE` - Scheduled
  - `mdtask/status/DONE` - Completed
- **Additional Tags**:
  - `mdtask/archived` - For archived tasks
  - `mdtask/deadline/YYYY-MM-DD` - For deadlines
  - `mdtask/waitfor/[reason]` - For wait reasons

### Planned Components

1. **CLI Application**: Full-featured command-line interface for task management
2. **Web Interface**: Browser-based UI launched via subcommand
3. **Configuration System**: TOML-based configuration for project directories

## Development Commands

Since this is a new Go project, when implementation begins:

```bash
# Initialize Go module (if not already done)
go mod init github.com/username/mdtask

# Run the application
go run main.go

# Build the application
go build -o mdtask

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run
```

## Key Design Decisions

1. **Markdown as Database**: Tasks are stored as individual Markdown files, making them human-readable and version-control friendly
2. **Tag-based Status**: Task status and metadata are managed through a hierarchical tag system
3. **Timestamp-based IDs**: Unique IDs are generated from creation timestamps to ensure uniqueness without external dependencies

## Git Commit Guidelines

- Do not add Co-Authored-By entries for Claude in commit messages
- Do not add "Generated with Claude Code" lines in commit messages
- Keep commit messages concise and descriptive