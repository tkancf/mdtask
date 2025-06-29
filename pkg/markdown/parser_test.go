package markdown

import (
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

func TestParseTaskFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *task.Task
		wantErr bool
	}{
		{
			name: "valid task with full timestamp",
			content: `---
id: task/20250101120000
aliases: []
tags:
    - mdtask
    - mdtask/status/TODO
    - project/test
created: 2025-01-01 12:00
description: Test description
title: Test Task
updated: 2025-01-01 13:00
---

This is the task content.
Multiple lines are supported.`,
			want: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Test Task",
				Description: "Test description",
				Aliases:     []string{},
				Tags:        []string{"mdtask", "mdtask/status/TODO", "project/test"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC),
				Content:     "This is the task content.\nMultiple lines are supported.",
			},
			wantErr: false,
		},
		{
			name: "valid task with date only",
			content: `---
id: task/20250101000000
aliases: []
tags:
    - mdtask
created: 2025-01-01
description: Date only test
title: Date Test
updated: 2025-01-02
---

Content here`,
			want: &task.Task{
				ID:          "task/20250101000000",
				Title:       "Date Test",
				Description: "Date only test",
				Aliases:     []string{},
				Tags:        []string{"mdtask"},
				Created:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				Content:     "Content here",
			},
			wantErr: false,
		},
		{
			name: "task with empty content",
			content: `---
id: task/20250101120000
aliases: []
tags:
    - mdtask
created: 2025-01-01 12:00
description: Empty content test
title: Empty Task
updated: 2025-01-01 12:00
---`,
			want: &task.Task{
				ID:          "task/20250101120000",
				Title:       "Empty Task",
				Description: "Empty content test",
				Aliases:     []string{},
				Tags:        []string{"mdtask"},
				Created:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Updated:     time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Content:     "",
			},
			wantErr: false,
		},
		{
			name:    "missing front matter",
			content: "Just some content without front matter",
			want:    nil,
			wantErr: true,
		},
		{
			name: "unclosed front matter",
			content: `---
id: task/20250101120000
title: Unclosed
created: 2025-01-01 12:00

This should fail`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid YAML",
			content: `---
id: task/20250101120000
tags:
  - mdtask
  invalid yaml here
---`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskFile([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				// Compare fields
				if got.ID != tt.want.ID {
					t.Errorf("ParseTaskFile() ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Title != tt.want.Title {
					t.Errorf("ParseTaskFile() Title = %v, want %v", got.Title, tt.want.Title)
				}
				if got.Description != tt.want.Description {
					t.Errorf("ParseTaskFile() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Content != tt.want.Content {
					t.Errorf("ParseTaskFile() Content = %v, want %v", got.Content, tt.want.Content)
				}
				if len(got.Tags) != len(tt.want.Tags) {
					t.Errorf("ParseTaskFile() Tags length = %v, want %v", len(got.Tags), len(tt.want.Tags))
				} else {
					for i, tag := range got.Tags {
						if tag != tt.want.Tags[i] {
							t.Errorf("ParseTaskFile() Tags[%d] = %v, want %v", i, tag, tt.want.Tags[i])
						}
					}
				}
			}
		})
	}
}

func TestExtractFrontMatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantFM      string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "valid front matter",
			content: `---
id: test
title: Test
---

Body content`,
			wantFM: `id: test
title: Test`,
			wantBody: "Body content",
			wantErr:  false,
		},
		{
			name: "multiple newlines after front matter",
			content: `---
id: test
---


Body content`,
			wantFM:   "id: test",
			wantBody: "Body content",
			wantErr:  false,
		},
		{
			name:     "no front matter",
			content:  "Just content",
			wantFM:   "",
			wantBody: "",
			wantErr:  true,
		},
		{
			name: "empty front matter",
			content: `---
---
Body`,
			wantFM:   "",
			wantBody: "Body",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFM, gotBody, err := extractFrontMatter([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFrontMatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFM != tt.wantFM {
				t.Errorf("extractFrontMatter() front matter = %v, want %v", gotFM, tt.wantFM)
			}
			if gotBody != tt.wantBody {
				t.Errorf("extractFrontMatter() body = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestBytesToStrings(t *testing.T) {
	tests := []struct {
		name  string
		input [][]byte
		want  []string
	}{
		{
			name:  "empty slice",
			input: [][]byte{},
			want:  []string{},
		},
		{
			name:  "single element",
			input: [][]byte{[]byte("hello")},
			want:  []string{"hello"},
		},
		{
			name:  "multiple elements",
			input: [][]byte{[]byte("hello"), []byte("world"), []byte("!")},
			want:  []string{"hello", "world", "!"},
		},
		{
			name:  "with empty strings",
			input: [][]byte{[]byte(""), []byte("test"), []byte("")},
			want:  []string{"", "test", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesToStrings(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("bytesToStrings() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("bytesToStrings()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}