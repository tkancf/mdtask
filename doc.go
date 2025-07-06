// Package mdtask provides a command-line task management tool that uses Markdown files as task tickets.
//
// mdtask treats each Markdown file as an individual task, storing metadata in YAML frontmatter
// and task content in the Markdown body. This approach makes tasks human-readable, version-control
// friendly, and easily editable with any text editor.
//
// Features:
//
//   - Task creation and management via CLI
//   - Tag-based task organization and status tracking
//   - Deadline and reminder support
//   - Hierarchical task relationships (parent/subtask)
//   - Full-text and tag-based search
//   - Task archiving
//   - JSON output for integration with other tools
//   - Neovim plugin for enhanced editing experience
//   - Web interface for visual task management
//
// Task Structure:
//
// Each task is stored as a Markdown file with the following structure:
//
//	---
//	id: task/20240101120000
//	title: Task Title
//	description: Brief description
//	tags:
//	  - mdtask
//	  - mdtask/status/TODO
//	  - custom-tag
//	created: 2024-01-01T12:00:00Z
//	updated: 2024-01-01T12:00:00Z
//	---
//	
//	Task content in Markdown format
//
// Status Management:
//
// Tasks can have one of the following statuses:
//   - TODO: Task not yet started
//   - WIP: Work in progress
//   - WAIT: Waiting for external input
//   - SCHE: Scheduled for future
//   - DONE: Completed
//
// File Organization:
//
// Task files are named using their creation timestamp (YYYYMMDDHHMMSS.md) and can be
// organized in any directory structure. The tool searches configured paths recursively
// to find all task files.
package main