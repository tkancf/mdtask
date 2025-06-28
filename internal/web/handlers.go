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
		
		t := &task.Task{
			ID:          markdown.GenerateTaskID(),
			Title:       title,
			Description: r.FormValue("description"),
			Content:     r.FormValue("content"),
			Created:     time.Now(),
			Updated:     time.Now(),
			Tags:        []string{"mdtask"},
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
		t.Updated = time.Now()

		// Find and update the file
		// This is simplified - in production, we'd track file paths
		http.Error(w, "Not implemented", http.StatusNotImplemented)

	case "DELETE":
		// Archive task
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

func (s *Server) render(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := s.templates.ExecuteTemplate(w, tmpl, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
	}
}