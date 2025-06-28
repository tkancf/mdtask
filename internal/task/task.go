package task

import (
	"time"
)

type Status string

const (
	StatusTODO Status = "TODO"
	StatusWIP  Status = "WIP"
	StatusWAIT Status = "WAIT"
	StatusSCHE Status = "SCHE"
	StatusDONE Status = "DONE"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Aliases     []string
	Tags        []string
	Created     time.Time
	Updated     time.Time
	Content     string
}

func (t *Task) GetStatus() Status {
	for _, tag := range t.Tags {
		switch tag {
		case "mdtask/status/TODO":
			return StatusTODO
		case "mdtask/status/WIP":
			return StatusWIP
		case "mdtask/status/WAIT":
			return StatusWAIT
		case "mdtask/status/SCHE":
			return StatusSCHE
		case "mdtask/status/DONE":
			return StatusDONE
		}
	}
	return StatusTODO
}

func (t *Task) SetStatus(status Status) {
	statusTags := []string{
		"mdtask/status/TODO",
		"mdtask/status/WIP",
		"mdtask/status/WAIT",
		"mdtask/status/SCHE",
		"mdtask/status/DONE",
	}
	
	newTags := []string{}
	for _, tag := range t.Tags {
		isStatusTag := false
		for _, statusTag := range statusTags {
			if tag == statusTag {
				isStatusTag = true
				break
			}
		}
		if !isStatusTag {
			newTags = append(newTags, tag)
		}
	}
	
	newTags = append(newTags, "mdtask/status/"+string(status))
	t.Tags = newTags
}

func (t *Task) IsArchived() bool {
	for _, tag := range t.Tags {
		if tag == "mdtask/archived" {
			return true
		}
	}
	return false
}

func (t *Task) Archive() {
	if !t.IsArchived() {
		t.Tags = append(t.Tags, "mdtask/archived")
	}
}

func (t *Task) GetDeadline() *time.Time {
	for _, tag := range t.Tags {
		if len(tag) > 16 && tag[:16] == "mdtask/deadline/" {
			dateStr := tag[16:]
			if deadline, err := time.Parse("2006-01-02", dateStr); err == nil {
				return &deadline
			}
		}
	}
	return nil
}

func (t *Task) SetDeadline(deadline time.Time) {
	newTags := []string{}
	for _, tag := range t.Tags {
		if len(tag) < 16 || tag[:16] != "mdtask/deadline/" {
			newTags = append(newTags, tag)
		}
	}
	
	deadlineTag := "mdtask/deadline/" + deadline.Format("2006-01-02")
	newTags = append(newTags, deadlineTag)
	t.Tags = newTags
}

func (t *Task) GetWaitReason() string {
	for _, tag := range t.Tags {
		if len(tag) > 15 && tag[:15] == "mdtask/waitfor/" {
			return tag[15:]
		}
	}
	return ""
}

func (t *Task) SetWaitReason(reason string) {
	newTags := []string{}
	for _, tag := range t.Tags {
		if len(tag) < 15 || tag[:15] != "mdtask/waitfor/" {
			newTags = append(newTags, tag)
		}
	}
	
	if reason != "" {
		waitTag := "mdtask/waitfor/" + reason
		newTags = append(newTags, waitTag)
	}
	t.Tags = newTags
}

func (t *Task) IsManagedTask() bool {
	for _, tag := range t.Tags {
		if tag == "mdtask" {
			return true
		}
	}
	return false
}