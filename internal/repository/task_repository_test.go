package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tkan/mdtask/internal/task"
)

func setupTestRepo(t *testing.T) (*TaskRepository, string) {
	tempDir, err := os.MkdirTemp("", "mdtask-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	repo := NewTaskRepository([]string{tempDir})
	return repo, tempDir
}

func createTestTask(t *testing.T, repo *TaskRepository, id, title string, tags []string) *task.Task {
	testTask := &task.Task{
		ID:          id,
		Title:       title,
		Description: "Test description",
		Tags:        tags,
		Created:     time.Now(),
		Updated:     time.Now(),
		Content:     "Test content",
	}
	
	_, err := repo.Create(testTask)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	
	return testTask
}

func TestNewTaskRepository(t *testing.T) {
	paths := []string{"/path1", "/path2"}
	repo := NewTaskRepository(paths)
	
	if len(repo.rootPaths) != len(paths) {
		t.Errorf("NewTaskRepository() paths length = %v, want %v", len(repo.rootPaths), len(paths))
	}
}

func TestCreate(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{
			name: "create new task with ID",
			task: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Test Task",
				Description: "Test",
				Tags:        []string{"project/test"},
				Created:     time.Now(),
				Updated:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "create task without ID",
			task: &task.Task{
				Title:       "No ID Task",
				Description: "Test",
				Tags:        []string{},
				Created:     time.Now(),
				Updated:     time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := repo.Create(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Create() file not created at %s", path)
				}
				
				// Verify task has mdtask tag
				hasMdtask := false
				for _, tag := range tt.task.Tags {
					if tag == "mdtask" {
						hasMdtask = true
						break
					}
				}
				if !hasMdtask {
					t.Errorf("Create() task should have mdtask tag")
				}
				
				// Verify task has status
				if tt.task.GetStatus() == "" {
					t.Errorf("Create() task should have a status")
				}
			}
		})
	}
}

func TestFindAll(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	// Create test tasks
	task1 := createTestTask(t, repo, "task/20250101120000", "Task 1", []string{"mdtask", "mdtask/status/TODO"})
	task2 := createTestTask(t, repo, "task/20250101130000", "Task 2", []string{"mdtask", "mdtask/status/DONE"})
	
	// Create a non-mdtask file
	nonTaskContent := `---
id: note/20250101140000
title: Not a task
tags:
    - note
---
Content`
	nonTaskPath := filepath.Join(tempDir, "note.md")
	os.WriteFile(nonTaskPath, []byte(nonTaskContent), 0644)

	tasks, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("FindAll() returned %d tasks, want 2", len(tasks))
	}

	// Verify we got the right tasks
	foundIDs := make(map[string]bool)
	for _, task := range tasks {
		foundIDs[task.ID] = true
	}

	if !foundIDs[task1.ID] || !foundIDs[task2.ID] {
		t.Errorf("FindAll() didn't find all expected tasks")
	}
}

func TestFindByID(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	task1 := createTestTask(t, repo, "task/20250101120000", "Task 1", []string{"mdtask", "mdtask/status/TODO"})
	createTestTask(t, repo, "task/20250101130000", "Task 2", []string{"mdtask", "mdtask/status/DONE"})

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "find existing task",
			id:      task1.ID,
			wantErr: false,
		},
		{
			name:    "find non-existent task",
			id:      "task/99999999999999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && found.ID != tt.id {
				t.Errorf("FindByID() returned task with ID %v, want %v", found.ID, tt.id)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	original := createTestTask(t, repo, "task/20250101120000", "Original Title", []string{"mdtask", "mdtask/status/TODO"})
	
	// Update the task
	original.Title = "Updated Title"
	original.Tags = append(original.Tags, "updated")
	
	err := repo.Update(original)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	updated, err := repo.FindByID(original.ID)
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Update() title = %v, want %v", updated.Title, "Updated Title")
	}

	hasUpdatedTag := false
	for _, tag := range updated.Tags {
		if tag == "updated" {
			hasUpdatedTag = true
			break
		}
	}
	if !hasUpdatedTag {
		t.Errorf("Update() didn't preserve new tag")
	}
}

func TestFindByStatus(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	createTestTask(t, repo, "task/20250101120000", "TODO Task", []string{"mdtask", "mdtask/status/TODO"})
	createTestTask(t, repo, "task/20250101130000", "WIP Task", []string{"mdtask", "mdtask/status/WIP"})
	createTestTask(t, repo, "task/20250101140000", "DONE Task", []string{"mdtask", "mdtask/status/DONE"})

	tests := []struct {
		name   string
		status task.Status
		want   int
	}{
		{
			name:   "find TODO tasks",
			status: task.StatusTODO,
			want:   1,
		},
		{
			name:   "find WIP tasks",
			status: task.StatusWIP,
			want:   1,
		},
		{
			name:   "find DONE tasks",
			status: task.StatusDONE,
			want:   1,
		},
		{
			name:   "find WAIT tasks",
			status: task.StatusWAIT,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := repo.FindByStatus(tt.status)
			if err != nil {
				t.Errorf("FindByStatus() error = %v", err)
				return
			}
			
			if len(tasks) != tt.want {
				t.Errorf("FindByStatus() returned %d tasks, want %d", len(tasks), tt.want)
			}
		})
	}
}

func TestFindActive(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	createTestTask(t, repo, "task/20250101120000", "Active Task 1", []string{"mdtask", "mdtask/status/TODO"})
	createTestTask(t, repo, "task/20250101130000", "Active Task 2", []string{"mdtask", "mdtask/status/WIP"})
	createTestTask(t, repo, "task/20250101140000", "Archived Task", []string{"mdtask", "mdtask/status/DONE", "mdtask/archived"})

	active, err := repo.FindActive()
	if err != nil {
		t.Fatalf("FindActive() error = %v", err)
	}

	if len(active) != 2 {
		t.Errorf("FindActive() returned %d tasks, want 2", len(active))
	}

	// Verify no archived tasks
	for _, task := range active {
		if task.IsArchived() {
			t.Errorf("FindActive() returned archived task %s", task.ID)
		}
	}
}

func TestSearch(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	task1 := &task.Task{
		ID:          "task/20250101120000",
		Title:       "Fix bug in search",
		Description: "Search functionality broken",
		Tags:        []string{"mdtask", "mdtask/status/TODO", "bug"},
		Content:     "The search feature needs fixing",
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	repo.Create(task1)

	task2 := &task.Task{
		ID:          "task/20250101130000",
		Title:       "Add new feature",
		Description: "Implement new functionality",
		Tags:        []string{"mdtask", "mdtask/status/WIP", "enhancement"},
		Content:     "This is about adding something new",
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	repo.Create(task2)

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{
			name:  "search in title",
			query: "bug",
			want:  1,
		},
		{
			name:  "search in description",
			query: "broken",
			want:  1,
		},
		{
			name:  "search in content",
			query: "fixing",
			want:  1,
		},
		{
			name:  "search in tags",
			query: "enhancement",
			want:  1,
		},
		{
			name:  "search case insensitive",
			query: "SEARCH",
			want:  1,
		},
		{
			name:  "no results",
			query: "nonexistent",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.Search(tt.query)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}
			
			if len(results) != tt.want {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.want)
			}
		})
	}
}

func TestSearchByTags(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	createTestTask(t, repo, "task/20250101120000", "Task 1", []string{"mdtask", "type/bug", "priority/high"})
	createTestTask(t, repo, "task/20250101130000", "Task 2", []string{"mdtask", "type/feature", "priority/high"})
	createTestTask(t, repo, "task/20250101140000", "Task 3", []string{"mdtask", "type/bug", "priority/low"})
	createTestTask(t, repo, "task/20250101150000", "Archived", []string{"mdtask", "type/bug", "mdtask/archived"})

	tests := []struct {
		name        string
		includeTags []string
		excludeTags []string
		orMode      bool
		want        int
	}{
		{
			name:        "AND mode - bug AND high priority",
			includeTags: []string{"type/bug", "priority/high"},
			excludeTags: []string{},
			orMode:      false,
			want:        1,
		},
		{
			name:        "OR mode - bug OR feature",
			includeTags: []string{"type/bug", "type/feature"},
			excludeTags: []string{},
			orMode:      true,
			want:        3, // excludes archived
		},
		{
			name:        "exclude high priority",
			includeTags: []string{"type/bug"},
			excludeTags: []string{"priority/high"},
			orMode:      false,
			want:        1,
		},
		{
			name:        "no include tags",
			includeTags: []string{},
			excludeTags: []string{},
			orMode:      false,
			want:        3, // all non-archived
		},
		{
			name:        "case insensitive",
			includeTags: []string{"TYPE/BUG"},
			excludeTags: []string{},
			orMode:      false,
			want:        2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.SearchByTags(tt.includeTags, tt.excludeTags, tt.orMode)
			if err != nil {
				t.Errorf("SearchByTags() error = %v", err)
				return
			}
			
			if len(results) != tt.want {
				t.Errorf("SearchByTags() returned %d results, want %d", len(results), tt.want)
			}
		})
	}
}

func TestHasTag(t *testing.T) {
	tests := []struct {
		name      string
		tags      []string
		searchTag string
		want      bool
	}{
		{
			name:      "exact match",
			tags:      []string{"mdtask", "type/bug", "priority/high"},
			searchTag: "type/bug",
			want:      true,
		},
		{
			name:      "case insensitive match",
			tags:      []string{"mdtask", "type/bug", "priority/high"},
			searchTag: "TYPE/BUG",
			want:      true,
		},
		{
			name:      "no match",
			tags:      []string{"mdtask", "type/bug"},
			searchTag: "type/feature",
			want:      false,
		},
		{
			name:      "empty tags",
			tags:      []string{},
			searchTag: "mdtask",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasTag(tt.tags, tt.searchTag); got != tt.want {
				t.Errorf("hasTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindByIDWithPath(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	task1 := createTestTask(t, repo, "task/20250101120000", "Task 1", []string{"mdtask"})

	foundTask, foundPath, err := repo.FindByIDWithPath(task1.ID)
	if err != nil {
		t.Fatalf("FindByIDWithPath() error = %v", err)
	}

	if foundTask.ID != task1.ID {
		t.Errorf("FindByIDWithPath() task ID = %v, want %v", foundTask.ID, task1.ID)
	}

	if !strings.Contains(foundPath, "20250101120000") {
		t.Errorf("FindByIDWithPath() path doesn't contain timestamp: %v", foundPath)
	}

	// Test non-existent task
	_, _, err = repo.FindByIDWithPath("task/99999999999999")
	if err == nil {
		t.Errorf("FindByIDWithPath() should error for non-existent task")
	}
}