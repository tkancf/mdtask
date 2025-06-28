package task

import (
	"reflect"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want Status
	}{
		{
			name: "TODO status",
			tags: []string{"mdtask", "mdtask/status/TODO"},
			want: StatusTODO,
		},
		{
			name: "WIP status",
			tags: []string{"mdtask", "mdtask/status/WIP"},
			want: StatusWIP,
		},
		{
			name: "WAIT status",
			tags: []string{"mdtask", "mdtask/status/WAIT"},
			want: StatusWAIT,
		},
		{
			name: "SCHE status",
			tags: []string{"mdtask", "mdtask/status/SCHE"},
			want: StatusSCHE,
		},
		{
			name: "DONE status",
			tags: []string{"mdtask", "mdtask/status/DONE"},
			want: StatusDONE,
		},
		{
			name: "no status tag defaults to TODO",
			tags: []string{"mdtask", "project/test"},
			want: StatusTODO,
		},
		{
			name: "empty tags defaults to TODO",
			tags: []string{},
			want: StatusTODO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			if got := task.GetStatus(); got != tt.want {
				t.Errorf("GetStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetStatus(t *testing.T) {
	tests := []struct {
		name       string
		initialTags []string
		setStatus  Status
		wantTags   []string
	}{
		{
			name:       "set status on task without status",
			initialTags: []string{"mdtask", "project/test"},
			setStatus:  StatusWIP,
			wantTags:   []string{"mdtask", "project/test", "mdtask/status/WIP"},
		},
		{
			name:       "change existing status",
			initialTags: []string{"mdtask", "mdtask/status/TODO", "project/test"},
			setStatus:  StatusDONE,
			wantTags:   []string{"mdtask", "project/test", "mdtask/status/DONE"},
		},
		{
			name:       "preserve non-status tags",
			initialTags: []string{"mdtask", "mdtask/status/WIP", "priority/high", "mdtask/deadline/2025-01-01"},
			setStatus:  StatusWAIT,
			wantTags:   []string{"mdtask", "priority/high", "mdtask/deadline/2025-01-01", "mdtask/status/WAIT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.SetStatus(tt.setStatus)
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("SetStatus() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestIsArchived(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{
			name: "archived task",
			tags: []string{"mdtask", "mdtask/archived"},
			want: true,
		},
		{
			name: "not archived",
			tags: []string{"mdtask", "mdtask/status/TODO"},
			want: false,
		},
		{
			name: "empty tags",
			tags: []string{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			if got := task.IsArchived(); got != tt.want {
				t.Errorf("IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArchive(t *testing.T) {
	tests := []struct {
		name     string
		initialTags []string
		wantTags []string
	}{
		{
			name:     "archive unarchived task",
			initialTags: []string{"mdtask", "mdtask/status/DONE"},
			wantTags: []string{"mdtask", "mdtask/status/DONE", "mdtask/archived"},
		},
		{
			name:     "archive already archived task",
			initialTags: []string{"mdtask", "mdtask/archived"},
			wantTags: []string{"mdtask", "mdtask/archived"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.Archive()
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("Archive() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestGetDeadline(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want *time.Time
	}{
		{
			name: "task with deadline",
			tags: []string{"mdtask", "mdtask/deadline/2025-01-15"},
			want: func() *time.Time {
				d := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
				return &d
			}(),
		},
		{
			name: "task without deadline",
			tags: []string{"mdtask", "mdtask/status/TODO"},
			want: nil,
		},
		{
			name: "invalid deadline format",
			tags: []string{"mdtask", "mdtask/deadline/invalid"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			got := task.GetDeadline()
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetDeadline() = %v, want nil", got)
				}
			} else {
				if got == nil || !got.Equal(*tt.want) {
					t.Errorf("GetDeadline() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSetDeadline(t *testing.T) {
	deadline := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name        string
		initialTags []string
		wantTags    []string
	}{
		{
			name:        "set deadline on task without deadline",
			initialTags: []string{"mdtask", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO", "mdtask/deadline/2025-02-01"},
		},
		{
			name:        "replace existing deadline",
			initialTags: []string{"mdtask", "mdtask/deadline/2025-01-15"},
			wantTags:    []string{"mdtask", "mdtask/deadline/2025-02-01"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.SetDeadline(deadline)
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("SetDeadline() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestRemoveDeadline(t *testing.T) {
	tests := []struct {
		name        string
		initialTags []string
		wantTags    []string
	}{
		{
			name:        "remove existing deadline",
			initialTags: []string{"mdtask", "mdtask/deadline/2025-01-15", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO"},
		},
		{
			name:        "remove deadline when none exists",
			initialTags: []string{"mdtask", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.RemoveDeadline()
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("RemoveDeadline() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestGetWaitReason(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "task with wait reason",
			tags: []string{"mdtask", "mdtask/waitfor/customer-response"},
			want: "customer-response",
		},
		{
			name: "task without wait reason",
			tags: []string{"mdtask", "mdtask/status/WAIT"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			if got := task.GetWaitReason(); got != tt.want {
				t.Errorf("GetWaitReason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetWaitReason(t *testing.T) {
	tests := []struct {
		name        string
		initialTags []string
		reason      string
		wantTags    []string
	}{
		{
			name:        "set wait reason",
			initialTags: []string{"mdtask", "mdtask/status/WAIT"},
			reason:      "approval",
			wantTags:    []string{"mdtask", "mdtask/status/WAIT", "mdtask/waitfor/approval"},
		},
		{
			name:        "replace existing wait reason",
			initialTags: []string{"mdtask", "mdtask/waitfor/old-reason"},
			reason:      "new-reason",
			wantTags:    []string{"mdtask", "mdtask/waitfor/new-reason"},
		},
		{
			name:        "remove wait reason with empty string",
			initialTags: []string{"mdtask", "mdtask/waitfor/reason"},
			reason:      "",
			wantTags:    []string{"mdtask"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.SetWaitReason(tt.reason)
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("SetWaitReason() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestIsManagedTask(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{
			name: "managed task",
			tags: []string{"mdtask", "mdtask/status/TODO"},
			want: true,
		},
		{
			name: "not managed task",
			tags: []string{"project/test", "priority/high"},
			want: false,
		},
		{
			name: "empty tags",
			tags: []string{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			if got := task.IsManagedTask(); got != tt.want {
				t.Errorf("IsManagedTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReminder(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want *time.Time
	}{
		{
			name: "reminder with time",
			tags: []string{"mdtask", "mdtask/reminder/2025-01-15T14:30"},
			want: func() *time.Time {
				r := time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)
				return &r
			}(),
		},
		{
			name: "reminder date only",
			tags: []string{"mdtask", "mdtask/reminder/2025-01-15"},
			want: func() *time.Time {
				r := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
				return &r
			}(),
		},
		{
			name: "no reminder",
			tags: []string{"mdtask"},
			want: nil,
		},
		{
			name: "invalid reminder format",
			tags: []string{"mdtask", "mdtask/reminder/invalid"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.tags}
			got := task.GetReminder()
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetReminder() = %v, want nil", got)
				}
			} else {
				if got == nil || !got.Equal(*tt.want) {
					t.Errorf("GetReminder() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSetReminder(t *testing.T) {
	reminder := time.Date(2025, 2, 1, 10, 30, 0, 0, time.UTC)
	
	tests := []struct {
		name        string
		initialTags []string
		wantTags    []string
	}{
		{
			name:        "set reminder on task without reminder",
			initialTags: []string{"mdtask", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO", "mdtask/reminder/2025-02-01T10:30"},
		},
		{
			name:        "replace existing reminder",
			initialTags: []string{"mdtask", "mdtask/reminder/2025-01-15T14:00"},
			wantTags:    []string{"mdtask", "mdtask/reminder/2025-02-01T10:30"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.SetReminder(reminder)
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("SetReminder() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}

func TestRemoveReminder(t *testing.T) {
	tests := []struct {
		name        string
		initialTags []string
		wantTags    []string
	}{
		{
			name:        "remove existing reminder",
			initialTags: []string{"mdtask", "mdtask/reminder/2025-01-15T14:00", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO"},
		},
		{
			name:        "remove reminder when none exists",
			initialTags: []string{"mdtask", "mdtask/status/TODO"},
			wantTags:    []string{"mdtask", "mdtask/status/TODO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Tags: tt.initialTags}
			task.RemoveReminder()
			if !reflect.DeepEqual(task.Tags, tt.wantTags) {
				t.Errorf("RemoveReminder() tags = %v, want %v", task.Tags, tt.wantTags)
			}
		})
	}
}