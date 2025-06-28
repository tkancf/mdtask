package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tkan/mdtask/internal/config"
	"github.com/tkan/mdtask/internal/errors"
	"github.com/tkan/mdtask/internal/task"
	"github.com/tkan/mdtask/pkg/markdown"
)

type PageData struct {
	Title       string
	Tasks       []*task.Task
	Task        *task.Task
	TotalTasks  int
	ActiveTasks int
	TodoCount   int
	WipCount    int
	WaitCount   int
	DoneCount   int
	Query       string
	Status      string
	// Statistics
	CreatedToday   int
	CompletedToday int
	UpdatedToday   int
	OverdueTasks   int
	UpcomingTasks  int
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tasks, err := s.repo.FindActive()
	if err != nil {
		handleError(w, errors.InternalError("Failed to load tasks", err))
		return
	}

	// Get all tasks for statistics
	allTasks, _ := s.repo.FindAll()
	stats := calculateDashboardStats(allTasks)
	
	data := PageData{
		Title:          "Dashboard",
		Tasks:          tasks,
		TotalTasks:     len(allTasks),
		ActiveTasks:    len(tasks),
		TodoCount:      stats.TODO,
		WipCount:       stats.WIP,
		WaitCount:      stats.WAIT,
		DoneCount:      stats.DONE,
		CreatedToday:   stats.CreatedToday,
		CompletedToday: stats.CompletedToday,
		UpdatedToday:   stats.UpdatedToday,
		OverdueTasks:   stats.OverdueTasks,
		UpcomingTasks:  stats.UpcomingTasks,
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		handleError(w, errors.InternalError("Failed to render template", err))
	}
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []*task.Task
	var err error
	var title string

	// Check for tag filters
	includeTags := r.URL.Query()["tags"]
	excludeTags := r.URL.Query()["exclude"]
	orMode := r.URL.Query().Get("or") == "true"
	includeArchived := r.URL.Query().Get("archived") == "true"

	if len(includeTags) > 0 || len(excludeTags) > 0 {
		// Tag-based search
		tasks, err = s.repo.SearchByTags(includeTags, excludeTags, orMode)
		if err != nil {
			handleError(w, errors.InternalError("Failed to search tasks", err))
			return
		}
		
		// Include archived if requested
		if includeArchived {
			allTasks, _ := s.repo.FindAll()
			for _, t := range allTasks {
				if t.IsArchived() && matchesTags(t, includeTags, excludeTags, orMode) {
					tasks = append(tasks, t)
				}
			}
		}
		
		title = fmt.Sprintf("Tasks filtered by tags (%d)", len(tasks))
	} else {
		// Default: show active tasks
		tasks, err = s.repo.FindActive()
		if err != nil {
			handleError(w, errors.InternalError("Failed to load tasks", err))
			return
		}
		title = "All Tasks"
	}

	data := PageData{
		Title: title,
		Tasks: tasks,
	}

	if err := s.templates.ExecuteTemplate(w, "tasks.html", data); err != nil {
		handleError(w, errors.InternalError("Failed to render template", err))
	}
}

func matchesTags(t *task.Task, includeTags, excludeTags []string, orMode bool) bool {
	// Check exclude tags first
	for _, excludeTag := range excludeTags {
		for _, tag := range t.Tags {
			if strings.EqualFold(tag, excludeTag) {
				return false
			}
		}
	}
	
	// Check include tags
	if len(includeTags) == 0 {
		return true
	}
	
	if orMode {
		// OR mode: at least one match
		for _, includeTag := range includeTags {
			for _, tag := range t.Tags {
				if strings.EqualFold(tag, includeTag) {
					return true
				}
			}
		}
		return false
	} else {
		// AND mode: all must match
		for _, includeTag := range includeTags {
			found := false
			for _, tag := range t.Tags {
				if strings.EqualFold(tag, includeTag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
}

func (s *Server) handleByStatus(w http.ResponseWriter, r *http.Request) {
	statusStr := strings.TrimPrefix(r.URL.Path, "/status/")
	
	var status task.Status
	switch statusStr {
	case "todo":
		status = task.StatusTODO
	case "wip":
		status = task.StatusWIP
	case "wait":
		status = task.StatusWAIT
	case "done":
		status = task.StatusDONE
	default:
		http.NotFound(w, r)
		return
	}

	tasks, err := s.repo.FindByStatus(status)
	if err != nil {
		handleError(w, errors.InternalError("Failed to load tasks", err))
		return
	}

	data := PageData{
		Title:  fmt.Sprintf("%s Tasks", status),
		Tasks:  tasks,
		Status: string(status),
	}

	if err := s.templates.ExecuteTemplate(w, "tasks.html", data); err != nil {
		handleError(w, errors.InternalError("Failed to render template", err))
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	tasks, err := s.repo.Search(query)
	if err != nil {
		handleError(w, errors.InternalError("Failed to search tasks", err))
		return
	}

	data := PageData{
		Title: fmt.Sprintf("Search Results for '%s'", query),
		Tasks: tasks,
		Query: query,
	}

	if err := s.templates.ExecuteTemplate(w, "tasks.html", data); err != nil {
		handleError(w, errors.InternalError("Failed to render template", err))
	}
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "New Task",
			Task:  &task.Task{},
		}
		if err := s.templates.ExecuteTemplate(w, "edit.html", data); err != nil {
			handleError(w, errors.InternalError("Failed to render template", err))
		}
		return
	}

	if r.Method == "POST" {
		// Reload config to get latest settings
		if cfg, err := config.LoadFromDefaultLocation(); err == nil {
			s.config = cfg
		}

		t := &task.Task{
			ID:      markdown.GenerateTaskID(),
			Created: time.Now(),
			Updated: time.Now(),
		}

		if err := s.parseTaskForm(r, t); err != nil {
			handleError(w, errors.ValidationError("form", "Invalid form data"))
			return
		}

		// Apply configuration
		s.applyTaskConfig(t)

		// Set default status if not set
		if t.GetStatus() == "" {
			t.SetStatus(task.StatusTODO)
		}

		// Create the task
		filePath, err := s.repo.Create(t)
		if err != nil {
			handleError(w, errors.InternalError("Failed to create task", err))
			return
		}

		fmt.Printf("Created task %s at %s\n", t.ID, filePath)
		http.Redirect(w, r, fmt.Sprintf("/task/%s", t.ID), http.StatusSeeOther)
	}
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/task/")
	
	t, err := s.repo.FindByID(id)
	if err != nil {
		handleError(w, err)
		return
	}

	data := PageData{
		Title: t.Title,
		Task:  t,
	}

	if err := s.templates.ExecuteTemplate(w, "task.html", data); err != nil {
		handleError(w, errors.InternalError("Failed to render template", err))
	}
}

func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/edit/")
	
	if r.Method == "GET" {
		t, err := s.repo.FindByID(id)
		if err != nil {
			handleError(w, err)
			return
		}

		data := PageData{
			Title: "Edit Task",
			Task:  t,
		}

		if err := s.templates.ExecuteTemplate(w, "edit.html", data); err != nil {
			handleError(w, errors.InternalError("Failed to render template", err))
		}
		return
	}

	if r.Method == "POST" {
		t, err := s.repo.FindByID(id)
		if err != nil {
			handleError(w, err)
			return
		}

		// Preserve mdtask-prefixed tags
		preservedTags := preserveMdtaskTags(t.Tags)

		// Parse form
		if err := s.parseTaskForm(r, t); err != nil {
			handleError(w, errors.ValidationError("form", "Invalid form data"))
			return
		}

		// Add back preserved tags
		t.Tags = append(t.Tags, preservedTags...)

		// Update task
		if err := s.repo.Update(t); err != nil {
			handleError(w, errors.InternalError("Failed to update task", err))
			return
		}

		fmt.Printf("Updated task %s\n", t.ID)
		http.Redirect(w, r, fmt.Sprintf("/task/%s", t.ID), http.StatusSeeOther)
	}
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/archive/")
	
	t, err := s.repo.FindByID(id)
	if err != nil {
		handleError(w, err)
		return
	}

	t.Archive()
	
	if err := s.repo.Update(t); err != nil {
		handleError(w, errors.InternalError("Failed to archive task", err))
		return
	}

	fmt.Printf("Archived task %s\n", t.ID)
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

// API handlers

type APITaskResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Content     string    `json:"content,omitempty"`
}

func (s *Server) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := s.repo.FindActive()
	if err != nil {
		handleError(w, errors.InternalError("Failed to load tasks", err))
		return
	}

	response := make([]APITaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = APITaskResponse{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Status:      string(t.GetStatus()),
			Tags:        t.Tags,
			Created:     t.Created,
			Updated:     t.Updated,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAPITask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/task/")
	
	t, err := s.repo.FindByID(id)
	if err != nil {
		handleError(w, err)
		return
	}

	response := APITaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.GetStatus()),
		Tags:        t.Tags,
		Created:     t.Created,
		Updated:     t.Updated,
		Content:     t.Content,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}