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

```bash
# Build the application
go build -o mdtask

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Run the test script
./test.sh

# Start WebUI (default port 7000)
./mdtask web --paths .

# Start WebUI on specific port
./mdtask web --port 8080 --paths .

# Example configuration file (see mdtask.toml.example)
cp mdtask.toml.example .mdtask.toml
```

## Key Design Decisions

1. **Markdown as Database**: Tasks are stored as individual Markdown files, making them human-readable and version-control friendly
2. **Tag-based Status**: Task status and metadata are managed through a hierarchical tag system
3. **Timestamp-based IDs**: Unique IDs are generated from creation timestamps to ensure uniqueness without external dependencies
4. **Automatic Port Selection**: WebUI starts on port 7000 by default, but automatically tries next ports if occupied
5. **Configuration System**: TOML-based config files with hierarchical search (current dir → home dir) for customization

## Git Commit Guidelines

- Do not add Co-Authored-By entries for Claude in commit messages
- Do not add "Generated with Claude Code" lines in commit messages
- Keep commit messages concise and descriptive
- **Make git commits at appropriate times during task implementation**:
  - After completing each major feature or component
  - When finishing a logical unit of work (e.g., after implementing a new command, fixing a bug, adding a new page)
  - Before switching to a different part of the codebase
  - After making significant changes that work correctly
  - Use descriptive commit messages that explain what was implemented
- **Commit frequently in appropriate units**:
  - Each commit should represent one logical change
  - Don't mix unrelated changes in a single commit
  - Commit after each successfully implemented feature, even if small
  - Examples of appropriate commit units:
    - Adding a new command (e.g., "Add stats command for task statistics")
    - Fixing a specific bug (e.g., "Fix WebUI edit status preservation")
    - Adding a new feature to existing functionality (e.g., "Add reminder support to task structure")
    - Updating UI/templates (e.g., "Add statistics section to dashboard")
    - Refactoring code (e.g., "Refactor task filtering logic")
  - Avoid commits that are too large (touching many unrelated files) or too small (fixing typos unless critical)