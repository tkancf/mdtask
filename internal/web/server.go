package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tkan/mdtask/internal/repository"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	repo      *repository.TaskRepository
	templates *template.Template
	port      string
}

func NewServer(repo *repository.TaskRepository, port string) (*Server, error) {
	funcMap := template.FuncMap{
		"now": time.Now,
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"ne": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		"lt": func(a, b int64) bool {
			return a < b
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		repo:      repo,
		templates: tmpl,
		port:      port,
	}, nil
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/tasks", s.handleTasks)
	mux.HandleFunc("/task/", s.handleTaskDetail)
	mux.HandleFunc("/api/task/", s.handleTaskAPI)
	mux.HandleFunc("/new", s.handleNewTask)

	// Static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Start server
	addr := fmt.Sprintf(":%s", s.port)
	fmt.Printf("Starting web server on http://localhost%s\n", addr)
	
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}