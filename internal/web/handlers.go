package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tkan/mdtask/internal/config"
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
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tasks, err := s.repo.FindActive()
	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:       "mdtask - Dashboard",
		Tasks:       tasks,
		TotalTasks:  len(tasks),
		ActiveTasks: len(tasks),
	}

	// Count tasks by status
	for _, t := range tasks {
		switch t.GetStatus() {
		case task.StatusTODO:
			data.TodoCount++
		case task.StatusWIP:
			data.WipCount++
		case task.StatusWAIT:
			data.WaitCount++
		case task.StatusDONE:
			data.DoneCount++
		}
	}

	s.render(w, "index.html", data)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := r.URL.Query().Get("q")

	var tasks []*task.Task
	var err error

	if query != "" {
		tasks, err = s.repo.Search(query)
	} else if status != "" {
		tasks, err = s.repo.FindByStatus(task.Status(status))
	} else {
		tasks, err = s.repo.FindActive()
	}

	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:  "Tasks",
		Tasks:  tasks,
		Query:  query,
		Status: status,
	}

	s.render(w, "tasks.html", data)
}

func (s *Server) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/task/")
	if taskID == "" {
		http.NotFound(w, r)
		return
	}

	t, err := s.repo.FindByID(taskID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title: t.Title,
		Task:  t,
	}

	s.render(w, "task.html", data)
}

func (s *Server) handleNewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "New Task",
		}
		s.render(w, "new.html", data)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		
		// Load configuration
		cfg, err := config.LoadFromDefaultLocation()
		if err != nil {
			http.Error(w, "Failed to load config", http.StatusInternalServerError)
			return
		}
		
		// Apply title prefix
		title := r.FormValue("title")
		if cfg.Task.TitlePrefix != "" {
			title = cfg.Task.TitlePrefix + title
		}
		
		// Apply templates if form values are empty
		description := r.FormValue("description")
		if description == "" && cfg.Task.DescriptionTemplate != "" {
			description = cfg.Task.DescriptionTemplate
		}
		
		content := r.FormValue("content")
		if content == "" && cfg.Task.ContentTemplate != "" {
			content = cfg.Task.ContentTemplate
		}
		
		// Start with mdtask tag and default tags from config
		tags := []string{"mdtask"}
		tags = append(tags, cfg.Task.DefaultTags...)
		
		t := &task.Task{
			ID:          markdown.GenerateTaskID(),
			Title:       title,
			Description: description,
			Content:     content,
			Created:     time.Now(),
			Updated:     time.Now(),
			Tags:        tags,
			Aliases:     []string{},
		}

		// Set status with config default
		status := r.FormValue("status")
		if status == "" {
			status = cfg.Task.DefaultStatus
			if status == "" {
				status = "TODO"
			}
		}
		t.SetStatus(task.Status(status))

		// Set deadline
		deadlineStr := r.FormValue("deadline")
		if deadlineStr != "" {
			if deadline, err := time.Parse("2006-01-02", deadlineStr); err == nil {
				t.SetDeadline(deadline)
			}
		}

		// Add tags
		tagsStr := r.FormValue("tags")
		if tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					t.Tags = append(t.Tags, tag)
				}
			}
		}

		_, err = s.repo.Create(t)
		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/task/%s", t.ID), http.StatusSeeOther)
	}
}

func (s *Server) handleKanban(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.repo.FindActive()
	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:       "Kanban Board",
		Tasks:       tasks,
		TotalTasks:  len(tasks),
		ActiveTasks: len(tasks),
	}

	// Count tasks by status
	for _, t := range tasks {
		switch t.GetStatus() {
		case task.StatusTODO:
			data.TodoCount++
		case task.StatusWIP:
			data.WipCount++
		case task.StatusWAIT:
			data.WaitCount++
		case task.StatusDONE:
			data.DoneCount++
		}
	}

	s.render(w, "kanban.html", data)
}

func (s *Server) handleEditTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/edit/")
	if taskID == "" {
		http.NotFound(w, r)
		return
	}

	if r.Method == "GET" {
		t, err := s.repo.FindByID(taskID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		data := PageData{
			Title: "Edit Task",
			Task:  t,
		}
		s.render(w, "edit.html", data)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()

		t, err := s.repo.FindByID(taskID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Load configuration
		cfg, err := config.LoadFromDefaultLocation()
		if err != nil {
			http.Error(w, "Failed to load config", http.StatusInternalServerError)
			return
		}

		// Update task fields
		title := r.FormValue("title")
		if cfg.Task.TitlePrefix != "" && !strings.HasPrefix(title, cfg.Task.TitlePrefix) {
			title = cfg.Task.TitlePrefix + title
		}
		t.Title = title
		t.Description = r.FormValue("description")
		t.Content = r.FormValue("content")

		// First, preserve existing mdtask/ prefixed tags
		var preservedTags []string
		for _, tag := range t.Tags {
			if strings.HasPrefix(tag, "mdtask/") && !strings.HasPrefix(tag, "mdtask/status/") && !strings.HasPrefix(tag, "mdtask/deadline/") {
				preservedTags = append(preservedTags, tag)
			}
		}

		// Update tags - start with mdtask tag and preserved tags
		t.Tags = []string{"mdtask"}
		t.Tags = append(t.Tags, preservedTags...)

		// Add user-provided tags
		tagsStr := r.FormValue("tags")
		if tagsStr != "" {
			tags := strings.Split(tagsStr, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" && tag != "mdtask" && !strings.HasPrefix(tag, "mdtask/") {
					t.Tags = append(t.Tags, tag)
				}
			}
		}

		// Update status (this will add the appropriate mdtask/status/ tag)
		status := r.FormValue("status")
		if status != "" {
			t.SetStatus(task.Status(status))
		} else {
			// If no status is provided, preserve the existing status
			currentStatus := t.GetStatus()
			if currentStatus != "" {
				t.SetStatus(currentStatus)
			} else {
				// Only set to TODO if there's truly no existing status
				t.SetStatus(task.StatusTODO)
			}
		}

		// Update deadline (this will add the appropriate mdtask/deadline/ tag)
		deadlineStr := r.FormValue("deadline")
		if deadlineStr != "" {
			if deadline, err := time.Parse("2006-01-02", deadlineStr); err == nil {
				t.SetDeadline(deadline)
			}
		} else {
			// Clear deadline if empty
			t.RemoveDeadline()
		}

		err = s.repo.Update(t)
		if err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/task/%s", t.ID), http.StatusSeeOther)
	}
}

func (s *Server) handleTaskAPI(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/task/")
	
	switch r.Method {
	case "PUT":
		// Update task status
		var req struct {
			Status string `json:"status"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		t, err := s.repo.FindByID(taskID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		t.SetStatus(task.Status(req.Status))

		err = s.repo.Update(t)
		if err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case "DELETE":
		// Archive task
		t, err := s.repo.FindByID(taskID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		t.Archive()
		
		err = s.repo.Update(t)
		if err != nil {
			http.Error(w, "Failed to archive task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) render(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := s.templates.ExecuteTemplate(w, tmpl, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
	}
}