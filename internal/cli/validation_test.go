package cli

import (
	"testing"
	"time"

	"github.com/tkancf/mdtask/internal/task"
)

func TestNormalizeTaskID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "already normalized ID",
			input:   "task/20240101120000",
			want:    "task/20240101120000",
			wantErr: false,
		},
		{
			name:    "timestamp only",
			input:   "20240101120000",
			want:    "task/20240101120000",
			wantErr: false,
		},
		{
			name:    "timestamp with suffix",
			input:   "20240101120000_1",
			want:    "task/20240101120000_1",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid-id",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "partial timestamp",
			input:   "202401",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTaskID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeTaskID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeTaskID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    task.Status
		wantErr bool
	}{
		{
			name:    "valid TODO",
			input:   "TODO",
			want:    task.StatusTODO,
			wantErr: false,
		},
		{
			name:    "valid WIP",
			input:   "WIP",
			want:    task.StatusWIP,
			wantErr: false,
		},
		{
			name:    "valid WAIT",
			input:   "WAIT",
			want:    task.StatusWAIT,
			wantErr: false,
		},
		{
			name:    "valid SCHE",
			input:   "SCHE",
			want:    task.StatusSCHE,
			wantErr: false,
		},
		{
			name:    "valid DONE",
			input:   "DONE",
			want:    task.StatusDONE,
			wantErr: false,
		},
		{
			name:    "invalid status",
			input:   "INVALID",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "lowercase",
			input:   "todo",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDeadline(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *time.Time
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   "2024-01-01",
			want:    timePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "empty string returns nil",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "01/01/2024",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   "2024-13-01",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "incomplete date",
			input:   "2024-01",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDeadline(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDeadline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalTimePtr(got, tt.want) {
				t.Errorf("ParseDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReminder(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *time.Time
		wantErr bool
	}{
		{
			name:    "valid datetime",
			input:   "2024-01-01 14:30",
			want:    timePtr(time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "empty string returns nil",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "2024-01-01",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid time",
			input:   "2024-01-01 25:00",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wrong separator",
			input:   "2024-01-01T14:30",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReminder(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReminder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalTimePtr(got, tt.want) {
				t.Errorf("ParseReminder() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func equalTimePtr(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}