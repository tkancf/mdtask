package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

func TestWriteTaskFile(t *testing.T) {
	tests := []struct {
		name    string
		task    *task.Task
		wantContains []string
		wantErr bool
	}{
		{
			name: "complete task",
			task: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Test Task",
				Description: "Test description",
				Aliases:     []string{"alias1", "alias2"},
				Tags:        []string{"mdtask", "mdtask/status/TODO", "project/test"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
				Content:     "This is the task content.\nMultiple lines are supported.",
			},
			wantContains: []string{
				"---",
				"id: task/20250101120000",
				"title: Test Task",
				"description: Test description",
				"aliases:",
				"    - alias1",
				"    - alias2",
				"tags:",
				"    - mdtask",
				"    - mdtask/status/TODO",
				"    - project/test",
				"created: 2025-01-01 12:00",
				"updated: 2025-01-01 13:00",
				"This is the task content.",
				"Multiple lines are supported.",
			},
			wantErr: false,
		},
		{
			name: "task with empty aliases",
			task: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Empty Aliases",
				Description: "Test",
				Aliases:     []string{},
				Tags:        []string{"mdtask"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Content:     "Content",
			},
			wantContains: []string{
				"aliases: []",
				"Content",
			},
			wantErr: false,
		},
		{
			name: "task with empty content",
			task: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Empty Content",
				Description: "Test",
				Aliases:     []string{},
				Tags:        []string{"mdtask"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Content:     "",
			},
			wantContains: []string{
				"---",
				"id: task/20250101120000",
				"title: Empty Content",
				"---",
			},
			wantErr: false,
		},
		{
			name: "task with special characters in title",
			task: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Test: Special Characters & \"Quotes\"",
				Description: "Test with 'quotes' and special chars",
				Aliases:     []string{},
				Tags:        []string{"mdtask"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Content:     "Content",
			},
			wantContains: []string{
				"title: 'Test: Special Characters & \"Quotes\"'",
				"description: Test with 'quotes' and special chars",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WriteTaskFile(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTaskFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				gotStr := string(got)
				for _, want := range tt.wantContains {
					if !strings.Contains(gotStr, want) {
						t.Errorf("WriteTaskFile() output missing %q\nGot:\n%s", want, gotStr)
					}
				}
				
				// Check structure
				if !strings.HasPrefix(gotStr, "---\n") {
					t.Errorf("WriteTaskFile() should start with ---")
				}
				if strings.Count(gotStr, "---") < 2 {
					t.Errorf("WriteTaskFile() should have at least two --- markers")
				}
			}
		})
	}
}

func TestGenerateTaskID(t *testing.T) {
	// Test basic ID generation
	id1 := GenerateTaskID()
	if !strings.HasPrefix(id1, "task/") {
		t.Errorf("GenerateTaskID() should start with 'task/', got %v", id1)
	}
	
	// Check format: task/YYYYMMDDHHMMSS
	parts := strings.Split(id1, "/")
	if len(parts) != 2 {
		t.Errorf("GenerateTaskID() should have format task/timestamp, got %v", id1)
	}
	
	if len(parts[1]) != 14 {
		t.Errorf("GenerateTaskID() timestamp should be 14 characters (YYYYMMDDHHMMSS), got %v", parts[1])
	}
	
	// Test uniqueness - generate two IDs
	id2 := GenerateTaskID()
	if id1 == id2 {
		t.Errorf("GenerateTaskID() should generate unique IDs, got duplicate: %v", id1)
	}
	
	// Verify they're different
	if id2 <= id1 {
		t.Errorf("GenerateTaskID() second ID should be greater than first, got %v <= %v", id2, id1)
	}
}

func TestGenerateTaskID_Concurrent(t *testing.T) {
	// Test concurrent ID generation
	ids := make(chan string, 3)
	
	for i := 0; i < 3; i++ {
		go func() {
			ids <- GenerateTaskID()
		}()
	}
	
	id1 := <-ids
	id2 := <-ids
	id3 := <-ids
	
	// All IDs should be unique
	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Errorf("GenerateTaskID() should generate unique IDs in concurrent calls, got: %v, %v, %v", id1, id2, id3)
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can write a task and parse it back
	original := &task.Task{
		ID:          "task/20250101120000",
		Title:       "Round Trip Test",
		Description: "Testing write and parse",
		Aliases:     []string{"test-alias"},
		Tags:        []string{"mdtask", "mdtask/status/WIP", "test"},
		Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Updated:     time.Date(2025, 1, 1, 13, 30, 0, 0, time.UTC),
		Content:     "This is test content.\n\nWith multiple paragraphs.",
	}
	
	// Write the task
	data, err := WriteTaskFile(original)
	if err != nil {
		t.Fatalf("WriteTaskFile() error = %v", err)
	}
	
	// Parse it back
	parsed, err := ParseTaskFile(data)
	if err != nil {
		t.Fatalf("ParseTaskFile() error = %v", err)
	}
	
	// Compare fields
	if parsed.ID != original.ID {
		t.Errorf("Round trip ID = %v, want %v", parsed.ID, original.ID)
	}
	if parsed.Title != original.Title {
		t.Errorf("Round trip Title = %v, want %v", parsed.Title, original.Title)
	}
	if parsed.Description != original.Description {
		t.Errorf("Round trip Description = %v, want %v", parsed.Description, original.Description)
	}
	// The writer adds a trailing newline, but parser removes it
	// Both behaviors are correct, so we should normalize for comparison
	originalNormalized := strings.TrimRight(original.Content, "\n")
	parsedNormalized := strings.TrimRight(parsed.Content, "\n")
	if parsedNormalized != originalNormalized {
		t.Errorf("Round trip Content = %q, want %q", parsed.Content, original.Content)
	}
	
	// Compare slices
	if len(parsed.Tags) != len(original.Tags) {
		t.Errorf("Round trip Tags length = %v, want %v", len(parsed.Tags), len(original.Tags))
	}
	if len(parsed.Aliases) != len(original.Aliases) {
		t.Errorf("Round trip Aliases length = %v, want %v", len(parsed.Aliases), len(original.Aliases))
	}
}