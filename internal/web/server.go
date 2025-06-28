package web

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
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
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
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
	mux.HandleFunc("/kanban", s.handleKanban)
	mux.HandleFunc("/task/", s.handleTaskDetail)
	mux.HandleFunc("/api/task/", s.handleTaskAPI)
	mux.HandleFunc("/new", s.handleNewTask)
	mux.HandleFunc("/edit/", s.handleEditTask)

	// Static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Try to start server on the specified port, if it fails, try the next few ports
	port := s.port
	maxRetries := 10
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Try next port
			nextPort, _ := strconv.Atoi(port)
			port = strconv.Itoa(nextPort + 1)
		}
		
		addr := fmt.Sprintf(":%s", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Successfully bound to port
		s.port = port // Update the actual port being used
		fmt.Printf("Starting web server on http://localhost:%s\n", port)
		
		server := &http.Server{
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		
		return server.Serve(listener)
	}
	
	return fmt.Errorf("could not start server after %d attempts: %w", maxRetries, lastErr)
}

// GetPort returns the actual port the server is running on
func (s *Server) GetPort() string {
	return s.port
}