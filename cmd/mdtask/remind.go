package mdtask

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/tkancf/mdtask/internal/repository"
)

var remindCmd = &cobra.Command{
	Use:   "remind",
	Short: "Check and show reminders for tasks",
	Long:  `Check all tasks for reminders and show notifications for tasks that are due.`,
	RunE:  runRemind,
}

var (
	remindCheck bool
	remindLoop  bool
)

func init() {
	rootCmd.AddCommand(remindCmd)
	remindCmd.Flags().BoolVarP(&remindCheck, "check", "c", false, "Check and list all tasks with reminders")
	remindCmd.Flags().BoolVarP(&remindLoop, "daemon", "d", false, "Run as daemon, checking reminders every minute")
}

func runRemind(cmd *cobra.Command, args []string) error {
	paths, _ := cmd.Flags().GetStringSlice("paths")
	repo := repository.NewTaskRepository(paths)

	if remindCheck {
		return checkReminders(repo)
	}

	if remindLoop {
		fmt.Println("Starting reminder daemon...")
		fmt.Println("Press Ctrl+C to stop")
		
		// Run once immediately
		if err := processReminders(repo); err != nil {
			fmt.Printf("Error processing reminders: %v\n", err)
		}
		
		// Then run every minute
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := processReminders(repo); err != nil {
				fmt.Printf("Error processing reminders: %v\n", err)
			}
		}
	} else {
		// Run once
		return processReminders(repo)
	}

	return nil
}

func checkReminders(repo *repository.TaskRepository) error {
	tasks, err := repo.FindActive()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	now := time.Now()
	hasReminders := false

	fmt.Println("Tasks with reminders:")
	fmt.Println()

	for _, task := range tasks {
		if reminder := task.GetReminder(); reminder != nil {
			hasReminders = true
			
			status := "Upcoming"
			if reminder.Before(now) {
				status = "Overdue"
			} else if reminder.Sub(now) < 24*time.Hour {
				status = "Due soon"
			}

			fmt.Printf("- [%s] %s\n", status, task.Title)
			fmt.Printf("  ID: %s\n", task.ID)
			fmt.Printf("  Reminder: %s\n", reminder.Format("2006-01-02 15:04"))
			fmt.Println()
		}
	}

	if !hasReminders {
		fmt.Println("No tasks with reminders found.")
	}

	return nil
}

func processReminders(repo *repository.TaskRepository) error {
	tasks, err := repo.FindActive()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	now := time.Now()
	
	for _, task := range tasks {
		if reminder := task.GetReminder(); reminder != nil {
			// Check if reminder is due (within the current minute)
			if reminder.Year() == now.Year() &&
				reminder.Month() == now.Month() &&
				reminder.Day() == now.Day() &&
				reminder.Hour() == now.Hour() &&
				reminder.Minute() == now.Minute() {
				
				if err := showNotification(task.Title, task.Description); err != nil {
					fmt.Printf("Failed to show notification for task %s: %v\n", task.ID, err)
				} else {
					fmt.Printf("Reminder shown for task: %s\n", task.Title)
				}
			}
		}
	}

	return nil
}

func showNotification(title, message string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("notifications are only supported on macOS")
	}

	// Use osascript to show macOS notification
	script := fmt.Sprintf(`display notification "%s" with title "mdtask reminder" subtitle "%s" sound name "default"`, 
		escapeForAppleScript(message), 
		escapeForAppleScript(title))
	
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func escapeForAppleScript(s string) string {
	// Escape quotes and backslashes for AppleScript
	result := ""
	for _, r := range s {
		switch r {
		case '"':
			result += "\\\""
		case '\\':
			result += "\\\\"
		default:
			result += string(r)
		}
	}
	return result
}