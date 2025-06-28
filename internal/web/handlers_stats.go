package web

import (
	"time"

	"github.com/tkan/mdtask/internal/constants"
	"github.com/tkan/mdtask/internal/task"
)

type DashboardStats struct {
	Total           int
	TODO            int
	WIP             int
	WAIT            int
	DONE            int
	SCHE            int
	CreatedToday    int
	CompletedToday  int
	UpdatedToday    int
	OverdueTasks    int
	UpcomingTasks   int
}

// calculateDashboardStats calculates statistics for the dashboard
func calculateDashboardStats(tasks []*task.Task) DashboardStats {
	stats := DashboardStats{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	for _, t := range tasks {
		if t.IsArchived() {
			continue
		}

		stats.Total++

		// Count by status
		switch t.GetStatus() {
		case task.StatusTODO:
			stats.TODO++
		case task.StatusWIP:
			stats.WIP++
		case task.StatusWAIT:
			stats.WAIT++
		case task.StatusSCHE:
			stats.SCHE++
		case task.StatusDONE:
			stats.DONE++
		}

		// Count today's activities
		if t.Created.After(todayStart) && t.Created.Before(todayEnd) {
			stats.CreatedToday++
		}
		if t.Updated.After(todayStart) && t.Updated.Before(todayEnd) {
			stats.UpdatedToday++
		}
		if t.GetStatus() == task.StatusDONE && t.Updated.After(todayStart) && t.Updated.Before(todayEnd) {
			stats.CompletedToday++
		}

		// Check deadlines
		if deadline := t.GetDeadline(); deadline != nil {
			if deadline.Before(now) {
				stats.OverdueTasks++
			} else if deadline.Sub(now) < constants.WeekDuration {
				stats.UpcomingTasks++
			}
		}
	}

	return stats
}