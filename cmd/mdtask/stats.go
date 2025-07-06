package mdtask

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/cli"
	"github.com/tkancf/mdtask/internal/task"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	Long:  `Display statistics about your tasks including daily progress, completion rates, and status breakdown.`,
	RunE:  runStats,
}

var (
	statsDate   string
	statsWeek   bool
	statsMonth  bool
	statsSimple bool
)

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().StringVarP(&statsDate, "date", "d", "", "Show stats for specific date (YYYY-MM-DD)")
	statsCmd.Flags().BoolVarP(&statsWeek, "week", "w", false, "Show stats for current week")
	statsCmd.Flags().BoolVarP(&statsMonth, "month", "m", false, "Show stats for current month")
	statsCmd.Flags().BoolVarP(&statsSimple, "simple", "s", false, "Show simple output without graphics")
}

type TaskStats struct {
	Total       int `json:"total"`
	ByStatus    struct {
		TODO int `json:"todo"`
		WIP  int `json:"wip"`
		WAIT int `json:"wait"`
		SCHE int `json:"sche"`
		DONE int `json:"done"`
	} `json:"by_status"`
	Activity struct {
		Created   int `json:"created"`
		Updated   int `json:"updated"`
		Completed int `json:"completed"`
	} `json:"activity"`
	Deadlines struct {
		Overdue  int `json:"overdue"`
		Upcoming int `json:"upcoming"`
	} `json:"deadlines"`
	DateRange struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"date_range"`
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx, err := cli.LoadContext(cmd)
	if err != nil {
		return err
	}

	tasks, err := ctx.Repo.FindAll()
	if err != nil {
		return err
	}

	// Determine the date range
	var startDate, endDate time.Time
	now := time.Now()
	
	if statsDate != "" {
		// Specific date
		startDate, err = time.Parse("2006-01-02", statsDate)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		endDate = startDate.AddDate(0, 0, 1)
	} else if statsWeek {
		// Current week (Monday to Sunday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = now.AddDate(0, 0, -(weekday - 1))
		endDate = startDate.AddDate(0, 0, 7)
	} else if statsMonth {
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	} else {
		// Today (default)
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 0, 1)
	}

	stats := calculateStats(tasks, startDate, endDate, now)
	
	if outputFormat == "json" {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(stats)
	}
	
	if statsSimple {
		displaySimpleStats(stats, startDate, endDate)
	} else {
		displayDetailedStats(stats, startDate, endDate, tasks)
	}

	return nil
}

func calculateStats(tasks []*task.Task, startDate, endDate, now time.Time) TaskStats {
	stats := TaskStats{}
	stats.DateRange.Start = startDate.Format("2006-01-02")
	stats.DateRange.End = endDate.AddDate(0, 0, -1).Format("2006-01-02")
	
	for _, t := range tasks {
		if t.IsArchived() {
			continue
		}
		
		stats.Total++
		
		// Count by status
		switch t.GetStatus() {
		case task.StatusTODO:
			stats.ByStatus.TODO++
		case task.StatusWIP:
			stats.ByStatus.WIP++
		case task.StatusWAIT:
			stats.ByStatus.WAIT++
		case task.StatusSCHE:
			stats.ByStatus.SCHE++
		case task.StatusDONE:
			stats.ByStatus.DONE++
		}
		
		// Check if created in date range
		if t.Created.After(startDate) && t.Created.Before(endDate) {
			stats.Activity.Created++
		}
		
		// Check if updated in date range
		if t.Updated.After(startDate) && t.Updated.Before(endDate) {
			stats.Activity.Updated++
		}
		
		// Check if completed in date range (assuming DONE tasks were completed when last updated)
		if t.GetStatus() == task.StatusDONE && t.Updated.After(startDate) && t.Updated.Before(endDate) {
			stats.Activity.Completed++
		}
		
		// Check deadlines
		if deadline := t.GetDeadline(); deadline != nil {
			if deadline.Before(now) {
				stats.Deadlines.Overdue++
			} else if deadline.Sub(now) < 7*24*time.Hour {
				stats.Deadlines.Upcoming++
			}
		}
	}
	
	return stats
}

func displaySimpleStats(stats TaskStats, startDate, endDate time.Time) {
	dateRange := formatDateRange(startDate, endDate)
	fmt.Printf("Task Statistics for %s\n", dateRange)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Total Active Tasks: %d\n", stats.Total)
	fmt.Printf("  TODO: %d\n", stats.ByStatus.TODO)
	fmt.Printf("  WIP:  %d\n", stats.ByStatus.WIP)
	fmt.Printf("  WAIT: %d\n", stats.ByStatus.WAIT)
	fmt.Printf("  SCHE: %d\n", stats.ByStatus.SCHE)
	fmt.Printf("  DONE: %d\n", stats.ByStatus.DONE)
	fmt.Println()
	fmt.Printf("Created:   %d\n", stats.Activity.Created)
	fmt.Printf("Updated:   %d\n", stats.Activity.Updated)
	fmt.Printf("Completed: %d\n", stats.Activity.Completed)
	fmt.Println()
	fmt.Printf("Overdue Tasks:      %d\n", stats.Deadlines.Overdue)
	fmt.Printf("Upcoming Deadlines: %d\n", stats.Deadlines.Upcoming)
}

func displayDetailedStats(stats TaskStats, startDate, endDate time.Time, tasks []*task.Task) {
	dateRange := formatDateRange(startDate, endDate)
	
	// Header
	fmt.Printf("\n📊 Task Statistics for %s\n", dateRange)
	fmt.Println(strings.Repeat("═", 60))
	
	// Progress Summary
	fmt.Println("\n📈 Progress Summary")
	fmt.Println(strings.Repeat("─", 40))
	
	if stats.Activity.Created > 0 || stats.Activity.Completed > 0 {
		fmt.Printf("✨ Created:   %d new task(s)\n", stats.Activity.Created)
		fmt.Printf("✅ Completed: %d task(s)\n", stats.Activity.Completed)
		fmt.Printf("📝 Updated:   %d task(s)\n", stats.Activity.Updated)
	} else {
		fmt.Println("No activity in this period")
	}
	
	// Status Breakdown with progress bars
	fmt.Println("\n📋 Status Breakdown")
	fmt.Println(strings.Repeat("─", 40))
	
	if stats.Total > 0 {
		displayStatusBar("TODO", stats.ByStatus.TODO, stats.Total, "🔵")
		displayStatusBar("WIP ", stats.ByStatus.WIP, stats.Total, "🟡")
		displayStatusBar("WAIT", stats.ByStatus.WAIT, stats.Total, "⚪")
		displayStatusBar("SCHE", stats.ByStatus.SCHE, stats.Total, "🟣")
		displayStatusBar("DONE", stats.ByStatus.DONE, stats.Total, "🟢")
		
		fmt.Printf("\nTotal Active Tasks: %d\n", stats.Total)
		
		// Completion rate for the period
		if stats.Activity.Created > 0 {
			completionRate := float64(stats.Activity.Completed) / float64(stats.Activity.Created) * 100
			fmt.Printf("\n🎯 Daily Completion Rate: %.1f%% (%d/%d)\n", 
				completionRate, stats.Activity.Completed, stats.Activity.Created)
		}
	} else {
		fmt.Println("No active tasks")
	}
	
	// Alerts
	if stats.Deadlines.Overdue > 0 || stats.Deadlines.Upcoming > 0 {
		fmt.Println("\n⚠️  Alerts")
		fmt.Println(strings.Repeat("─", 40))
		
		if stats.Deadlines.Overdue > 0 {
			fmt.Printf("🚨 Overdue Tasks: %d\n", stats.Deadlines.Overdue)
			// List overdue tasks
			for _, t := range tasks {
				if !t.IsArchived() && t.GetDeadline() != nil && t.GetDeadline().Before(time.Now()) {
					fmt.Printf("   - %s (due: %s)\n", t.Title, t.GetDeadline().Format("2006-01-02"))
				}
			}
		}
		
		if stats.Deadlines.Upcoming > 0 {
			fmt.Printf("\n📅 Upcoming Deadlines (next 7 days): %d\n", stats.Deadlines.Upcoming)
		}
	}
	
	// Tasks in progress
	if stats.ByStatus.WIP > 0 {
		fmt.Println("\n🚧 Tasks in Progress")
		fmt.Println(strings.Repeat("─", 40))
		for _, t := range tasks {
			if !t.IsArchived() && t.GetStatus() == task.StatusWIP {
				fmt.Printf("- %s\n", t.Title)
			}
		}
	}
	
	fmt.Println()
}

func displayStatusBar(label string, count, total int, emoji string) {
	if total == 0 {
		return
	}
	
	percentage := float64(count) / float64(total) * 100
	barLength := 20
	filled := int(float64(barLength) * float64(count) / float64(total))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLength-filled)
	
	fmt.Printf("%s %s [%s] %3d (%.1f%%)\n", emoji, label, bar, count, percentage)
}

func formatDateRange(startDate, endDate time.Time) string {
	duration := endDate.Sub(startDate)
	
	if duration.Hours() <= 24 {
		// Single day
		if startDate.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			return "Today"
		}
		return startDate.Format("2006-01-02")
	} else if duration.Hours() <= 24*7 {
		// Week
		return fmt.Sprintf("Week of %s", startDate.Format("2006-01-02"))
	} else if duration.Hours() <= 24*31 {
		// Month
		return startDate.Format("January 2006")
	}
	
	return fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.AddDate(0, 0, -1).Format("2006-01-02"))
}