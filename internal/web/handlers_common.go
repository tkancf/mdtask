package web

import (
	"net/http"
	"strings"

	"github.com/tkancf/mdtask/internal/constants"
	"github.com/tkancf/mdtask/internal/errors"
	"github.com/tkancf/mdtask/internal/task"
)

// parseFormTags extracts and normalizes tags from form input
func parseFormTags(formTags string, additionalTags []string) []string {
	tagsMap := make(map[string]bool)
	
	// Add form tags
	if formTags != "" {
		for _, tag := range strings.Split(formTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagsMap[tag] = true
			}
		}
	}
	
	// Add additional tags
	for _, tag := range additionalTags {
		if tag != "" {
			tagsMap[tag] = true
		}
	}
	
	// Convert map to slice
	tags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	
	return tags
}

// preserveMdtaskTags preserves mdtask-prefixed tags from the original task
func preserveMdtaskTags(originalTags []string) []string {
	var preserved []string
	for _, tag := range originalTags {
		if strings.HasPrefix(tag, constants.TagPrefix+"/") && 
		   !strings.HasPrefix(tag, constants.StatusTagPrefix) && 
		   !strings.HasPrefix(tag, constants.DeadlineTagPrefix) {
			preserved = append(preserved, tag)
		}
	}
	return preserved
}

// handleError sends an appropriate HTTP error response based on error type
func handleError(w http.ResponseWriter, err error) {
	if errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if errors.IsValidation(err) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if errors.IsConflict(err) {
		http.Error(w, err.Error(), http.StatusConflict)
	} else if errors.IsPermission(err) {
		http.Error(w, err.Error(), http.StatusForbidden)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getStatusFromForm gets the status from form value
func getStatusFromForm(formStatus string) task.Status {
	switch formStatus {
	case "WIP":
		return task.StatusWIP
	case "WAIT":
		return task.StatusWAIT
	case "SCHE":
		return task.StatusSCHE
	case "DONE":
		return task.StatusDONE
	default:
		return task.StatusTODO
	}
}