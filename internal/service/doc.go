// Package service provides the business logic layer for task operations.
//
// This package implements the core business rules and workflows for task management,
// separating them from the CLI command handlers and data access layer. It ensures
// consistent behavior across different interfaces (CLI, API, etc.).
//
// Key Features:
//
//   - Task creation with validation and defaults
//   - Task updates with field-level control
//   - Hierarchical task management (parent/subtask relationships)
//   - Task archiving with cascade operations
//   - Configuration-aware operations (templates, defaults)
//
// Architecture:
//
// The service layer sits between the CLI handlers and the repository layer:
//
//	CLI Commands -> Service Layer -> Repository -> File System
//
// This separation allows for:
//   - Easier testing of business logic
//   - Consistent validation and error handling
//   - Reusable operations across different interfaces
//   - Clear separation of concerns
//
// Example Usage:
//
//	service := service.NewTaskService(repo, config)
//	
//	// Create a new task
//	task, filePath, err := service.CreateTask(service.CreateTaskParams{
//	    Title:       "My Task",
//	    Description: "Task description",
//	    Status:      "TODO",
//	})
//	
//	// Archive a task and its subtasks
//	task, err := service.ArchiveTask(taskID)
package service