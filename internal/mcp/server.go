package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tkancf/mdtask/internal/config"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
)

type Server struct {
	repo   repository.Repository
	config *config.Config
	mcp    *server.MCPServer
}

func NewServer(repo repository.Repository, cfg *config.Config) *Server {
	s := &Server{
		repo:   repo,
		config: cfg,
	}

	// Create MCP server
	s.mcp = server.NewMCPServer(
		"mdtask MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// Register tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

func (s *Server) registerTools() {
	// List tasks tool
	listTool := mcp.NewTool("list_tasks",
		mcp.WithDescription("List all tasks or filter by status"),
		mcp.WithString("status",
			mcp.Description("Filter by status (TODO, WIP, WAIT, SCHE, DONE)"),
		),
		mcp.WithBoolean("archived",
			mcp.Description("Include archived tasks"),
		),
	)
	s.mcp.AddTool(listTool, s.listTasksHandler)

	// Create task tool
	createTool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Task title"),
		),
		mcp.WithString("description",
			mcp.Description("Task description"),
		),
		mcp.WithString("status",
			mcp.Description("Initial status (TODO, WIP, WAIT, SCHE, DONE)"),
		),
		mcp.WithArray("tags",
			mcp.Description("Additional tags for the task"),
		),
	)
	s.mcp.AddTool(createTool, s.createTaskHandler)

	// Update task tool
	updateTool := mcp.NewTool("update_task",
		mcp.WithDescription("Update an existing task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("title",
			mcp.Description("New title"),
		),
		mcp.WithString("description",
			mcp.Description("New description"),
		),
		mcp.WithString("status",
			mcp.Description("New status (TODO, WIP, WAIT, SCHE, DONE)"),
		),
		mcp.WithArray("add_tags",
			mcp.Description("Tags to add"),
		),
		mcp.WithArray("remove_tags",
			mcp.Description("Tags to remove"),
		),
	)
	s.mcp.AddTool(updateTool, s.updateTaskHandler)

	// Search tasks tool
	searchTool := mcp.NewTool("search_tasks",
		mcp.WithDescription("Search tasks by query"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithBoolean("archived",
			mcp.Description("Include archived tasks"),
		),
	)
	s.mcp.AddTool(searchTool, s.searchTasksHandler)

	// Archive task tool
	archiveTool := mcp.NewTool("archive_task",
		mcp.WithDescription("Archive a task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to archive"),
		),
	)
	s.mcp.AddTool(archiveTool, s.archiveTaskHandler)

	// Get task tool
	getTool := mcp.NewTool("get_task",
		mcp.WithDescription("Get details of a specific task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)
	s.mcp.AddTool(getTool, s.getTaskHandler)

	// Statistics tool
	statsTool := mcp.NewTool("get_statistics",
		mcp.WithDescription("Get task statistics"),
	)
	s.mcp.AddTool(statsTool, s.getStatisticsHandler)
}

func (s *Server) registerResources() {
	// Task list resource
	tasksResource := mcp.NewResource("tasks", "Active Tasks",
		mcp.WithResourceDescription("List of all active tasks"),
	)
	s.mcp.AddResource(tasksResource, s.tasksResourceHandler)

	// Statistics resource
	statsResource := mcp.NewResource("statistics", "Task Statistics",
		mcp.WithResourceDescription("Task statistics"),
	)
	s.mcp.AddResource(statsResource, s.statisticsResourceHandler)
}

func (s *Server) listTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := request.GetString("status", "")
	includeArchived := request.GetBool("archived", false)

	var tasks []*task.Task
	var err error

	if status != "" {
		tasks, err = s.repo.FindByStatus(task.Status(strings.ToUpper(status)))
	} else {
		tasks, err = s.repo.FindAll()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Filter out archived if needed
	if !includeArchived {
		filtered := make([]*task.Task, 0)
		for _, t := range tasks {
			if !t.IsArchived() {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Format response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d tasks\n\n", len(tasks)))
	
	for _, t := range tasks {
		status := string(t.GetStatus())
		
		result.WriteString(fmt.Sprintf("ID: %s\n", t.ID))
		result.WriteString(fmt.Sprintf("Title: %s\n", t.Title))
		result.WriteString(fmt.Sprintf("Status: %s\n", status))
		if t.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
		}
		result.WriteString(fmt.Sprintf("Created: %s\n", t.Created.Format("2006-01-02 15:04:05")))
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (s *Server) createTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := request.GetString("title", "")
	description := request.GetString("description", "")
	status := request.GetString("status", "")
	tagsRaw := request.GetStringSlice("tags", []string{})

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Create task
	t := &task.Task{
		Title:       title,
		Description: description,
		Tags:        []string{"mdtask"},
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	// Set status
	if status == "" {
		status = "TODO"
	}
	t.SetStatus(task.Status(strings.ToUpper(status)))

	// Add additional tags
	for _, tag := range tagsRaw {
		t.Tags = append(t.Tags, tag)
	}

	// Create in repository
	id, err := s.repo.Create(t)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	t.ID = id

	result := fmt.Sprintf("Task created successfully\nID: %s\nTitle: %s", t.ID, t.Title)
	return mcp.NewToolResultText(result), nil
}

func (s *Server) updateTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	// Get existing task
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update fields
	title := request.GetString("title", "")
	if title != "" {
		t.Title = title
	}
	description := request.GetString("description", "")
	if description != "" {
		t.Description = description
	}

	// Update status
	status := request.GetString("status", "")
	if status != "" {
		t.SetStatus(task.Status(strings.ToUpper(status)))
	}

	// Add tags
	addTags := request.GetStringSlice("add_tags", []string{})
	for _, tag := range addTags {
		t.Tags = append(t.Tags, tag)
	}

	// Remove tags
	removeTags := request.GetStringSlice("remove_tags", []string{})
	if len(removeTags) > 0 {
		removeSet := make(map[string]bool)
		for _, tag := range removeTags {
			removeSet[tag] = true
		}
		
		newTags := make([]string, 0)
		for _, tag := range t.Tags {
			if !removeSet[tag] {
				newTags = append(newTags, tag)
			}
		}
		t.Tags = newTags
	}

	t.Updated = time.Now()

	// Update in repository
	if err := s.repo.Update(t); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	result := fmt.Sprintf("Task updated successfully\nID: %s\nTitle: %s", t.ID, t.Title)
	return mcp.NewToolResultText(result), nil
}

func (s *Server) searchTasksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	includeArchived := request.GetBool("archived", false)

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	tasks, err := s.repo.Search(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	// Filter out archived if needed
	if !includeArchived {
		filtered := make([]*task.Task, 0)
		for _, t := range tasks {
			if !t.IsArchived() {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Format response
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d tasks matching '%s'\n\n", len(tasks), query))
	
	for _, t := range tasks {
		status := string(t.GetStatus())
		
		result.WriteString(fmt.Sprintf("ID: %s\n", t.ID))
		result.WriteString(fmt.Sprintf("Title: %s\n", t.Title))
		result.WriteString(fmt.Sprintf("Status: %s\n", status))
		if t.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (s *Server) archiveTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	// Get task
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Add archive tag
	if !t.IsArchived() {
		t.Archive()
		t.Updated = time.Now()
		
		if err := s.repo.Update(t); err != nil {
			return nil, fmt.Errorf("failed to archive task: %w", err)
		}
	}

	result := fmt.Sprintf("Task archived successfully\nID: %s\nTitle: %s", t.ID, t.Title)
	return mcp.NewToolResultText(result), nil
}

func (s *Server) getTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Format response
	var result strings.Builder
	
	status := string(t.GetStatus())
	
	result.WriteString(fmt.Sprintf("ID: %s\n", t.ID))
	result.WriteString(fmt.Sprintf("Title: %s\n", t.Title))
	result.WriteString(fmt.Sprintf("Status: %s\n", status))
	if t.Description != "" {
		result.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
	}
	result.WriteString(fmt.Sprintf("Created: %s\n", t.Created.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Updated: %s\n", t.Updated.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(t.Tags, ", ")))
	
	if t.Content != "" {
		result.WriteString("\nContent:\n")
		result.WriteString(t.Content)
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (s *Server) getStatisticsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tasks, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Calculate statistics
	stats := make(map[string]int)
	totalActive := 0
	totalArchived := 0

	for _, t := range tasks {
		if t.IsArchived() {
			totalArchived++
		} else {
			totalActive++
			
			// Count by status
			status := string(t.GetStatus())
			stats[status]++
		}
	}

	// Format response
	var result strings.Builder
	result.WriteString("Task Statistics\n\n")
	result.WriteString(fmt.Sprintf("Total tasks: %d\n", len(tasks)))
	result.WriteString(fmt.Sprintf("Active: %d\n", totalActive))
	result.WriteString(fmt.Sprintf("Archived: %d\n\n", totalArchived))
	
	result.WriteString("By Status:\n")
	for status, count := range stats {
		result.WriteString(fmt.Sprintf("  %s: %d\n", status, count))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (s *Server) tasksResourceHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	tasks, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Filter active tasks
	activeTasks := make([]*task.Task, 0)
	for _, t := range tasks {
		if !t.IsArchived() {
			activeTasks = append(activeTasks, t)
		}
	}

	// Format as markdown
	var content strings.Builder
	content.WriteString("# Active Tasks\n\n")
	
	// Group by status
	byStatus := make(map[string][]*task.Task)
	for _, t := range activeTasks {
		status := string(t.GetStatus())
		byStatus[status] = append(byStatus[status], t)
	}

	// Output by status
	for _, status := range []string{"TODO", "WIP", "WAIT", "SCHE", "DONE"} {
		if tasks, ok := byStatus[status]; ok && len(tasks) > 0 {
			content.WriteString(fmt.Sprintf("## %s (%d)\n\n", status, len(tasks)))
			for _, t := range tasks {
				content.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", t.Title, t.ID))
				if t.Description != "" {
					content.WriteString(fmt.Sprintf("  %s\n", t.Description))
				}
			}
			content.WriteString("\n")
		}
	}

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/markdown",
			Text:     content.String(),
		},
	}, nil
}

func (s *Server) statisticsResourceHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	tasks, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Calculate statistics
	stats := make(map[string]int)
	totalActive := 0
	totalArchived := 0

	for _, t := range tasks {
		if t.IsArchived() {
			totalArchived++
		} else {
			totalActive++
			
			// Count by status
			status := string(t.GetStatus())
			stats[status]++
		}
	}

	// Format as JSON
	jsonStats := fmt.Sprintf(`{
	"total": %d,
	"active": %d,
	"archived": %d,
	"byStatus": %v
}`, len(tasks), totalActive, totalArchived, stats)

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     jsonStats,
		},
	}, nil
}

func (s *Server) Serve() error {
	return server.ServeStdio(s.mcp)
}