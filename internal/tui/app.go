package tui

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tkancf/mdtask/internal/repository"
	"github.com/tkancf/mdtask/internal/task"
	"github.com/tkancf/mdtask/internal/tui/components"
	"github.com/tkancf/mdtask/internal/tui/views"
)

const (
	listHeight = 20
	listWidth  = 80
)

type viewState int

const (
	listView viewState = iota
	detailView
	statusSelectView
	createView
)

type undoAction struct {
	taskID     string
	oldStatus  task.Status
	newStatus  task.Status
	timestamp  time.Time
}

type App struct {
	repo           repository.Repository
	list           list.Model
	tasks          []*task.Task
	detail         *views.DetailView
	statusSelector *components.StatusSelector
	taskForm       *components.TaskForm
	selectedTask   *task.Task
	selectedTasks  map[string]*task.Task // For multi-select
	undoHistory    []undoAction          // History for undo
	viewState      viewState
	width          int
	height         int
	quitting       bool
	err            error
}

// taskItem implements list.Item interface
type taskItem struct {
	task     *task.Task
	selected bool
}

func (i taskItem) FilterValue() string { return i.task.Title }
func (i taskItem) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.task.Title
}
func (i taskItem) Description() string {
	status := ""
	for _, tag := range i.task.Tags {
		if tag == "mdtask/status/TODO" {
			status = "TODO"
			break
		} else if tag == "mdtask/status/WIP" {
			status = "WIP"
			break
		} else if tag == "mdtask/status/WAIT" {
			status = "WAIT"
			break
		} else if tag == "mdtask/status/SCHE" {
			status = "SCHE"
			break
		} else if tag == "mdtask/status/DONE" {
			status = "DONE"
			break
		}
	}
	return fmt.Sprintf("Status: %s | ID: %s", status, i.task.ID)
}

// itemDelegate is a custom delegate for list items
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 3 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(taskItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s\n   %s", index+1, i.Title(), i.Description())

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170")).
				Render("> " + s[0])
		}
	}

	fmt.Fprint(w, fn(str))
}

func NewApp(repo repository.Repository) *App {
	// Create list with custom delegate
	items := []list.Item{}
	delegate := itemDelegate{}
	l := list.New(items, delegate, listWidth, listHeight)
	l.Title = "mdtask - Task Manager"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Update key bindings
	l.KeyMap.ShowFullHelp.SetEnabled(true)
	l.KeyMap.Quit = key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "view task"),
			),
			key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new task"),
			),
			key.NewBinding(
				key.WithKeys("v"),
				key.WithHelp("v", "select/deselect"),
			),
			key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "change status"),
			),
			key.NewBinding(
				key.WithKeys("u"),
				key.WithHelp("u", "undo"),
			),
		}
	}

	return &App{
		repo:          repo,
		list:          l,
		selectedTasks: make(map[string]*task.Task),
		undoHistory:   make([]undoAction, 0),
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadTasks
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		if a.viewState == listView {
			a.list.SetWidth(msg.Width)
			a.list.SetHeight(msg.Height - 2)
		}
		return a, nil

	case tea.KeyMsg:
		switch a.viewState {
		case listView:
			switch msg.String() {
			case "q", "ctrl+c":
				a.quitting = true
				return a, tea.Quit
			case "enter":
				// Open task detail view
				if i, ok := a.list.SelectedItem().(taskItem); ok {
					a.detail = views.NewDetailView(i.task)
					a.selectedTask = i.task
					a.viewState = detailView
					return a, a.detail.Init()
				}
			case "v":
				// Toggle selection for multi-select
				if i, ok := a.list.SelectedItem().(taskItem); ok {
					if _, exists := a.selectedTasks[i.task.ID]; exists {
						delete(a.selectedTasks, i.task.ID)
					} else {
						a.selectedTasks[i.task.ID] = i.task
					}
					// Update the item in the list
					items := a.list.Items()
					for idx, item := range items {
						if ti, ok := item.(taskItem); ok && ti.task.ID == i.task.ID {
							ti.selected = !ti.selected
							items[idx] = ti
							break
						}
					}
					a.list.SetItems(items)
				}
			case "s":
				// Open status selector
				if len(a.selectedTasks) > 0 {
					// Bulk status change
					a.statusSelector = components.NewStatusSelector(task.StatusTODO)
					a.viewState = statusSelectView
					return a, a.statusSelector.Init()
				} else if i, ok := a.list.SelectedItem().(taskItem); ok {
					// Single task status change
					a.selectedTask = i.task
					a.statusSelector = components.NewStatusSelector(a.selectedTask.GetStatus())
					a.viewState = statusSelectView
					return a, a.statusSelector.Init()
				}
			case "a":
				// Select all tasks
				items := a.list.Items()
				for idx, item := range items {
					if ti, ok := item.(taskItem); ok {
						ti.selected = true
						a.selectedTasks[ti.task.ID] = ti.task
						items[idx] = ti
					}
				}
				a.list.SetItems(items)
			case "A":
				// Clear all selections
				a.selectedTasks = make(map[string]*task.Task)
				items := a.list.Items()
				for idx, item := range items {
					if ti, ok := item.(taskItem); ok {
						ti.selected = false
						items[idx] = ti
					}
				}
				a.list.SetItems(items)
			case "u":
				// Undo last action
				if len(a.undoHistory) > 0 {
					return a, a.undoLastAction()
				}
			case "n":
				// Create new task
				a.taskForm = components.NewTaskForm()
				a.viewState = createView
				return a, a.taskForm.Init()
			}
		}

	case tasksLoadedMsg:
		a.tasks = msg.tasks
		items := make([]list.Item, len(a.tasks))
		for i, t := range a.tasks {
			items[i] = taskItem{task: t}
		}
		a.list.SetItems(items)
		return a, nil

	case views.GoBackMsg:
		a.viewState = listView
		a.detail = nil
		return a, nil

	case components.StatusSelectedMsg:
		if len(a.selectedTasks) > 0 {
			// Bulk status update - record old statuses for undo
			for _, t := range a.selectedTasks {
				a.undoHistory = append(a.undoHistory, undoAction{
					taskID:    t.ID,
					oldStatus: t.GetStatus(),
					newStatus: msg.Status,
					timestamp: time.Now(),
				})
			}
			a.viewState = listView
			return a, a.updateTasks(a.selectedTasks, msg.Status)
		} else if a.selectedTask != nil {
			// Single task update - record for undo
			oldStatus := a.selectedTask.GetStatus()
			a.undoHistory = append(a.undoHistory, undoAction{
				taskID:    a.selectedTask.ID,
				oldStatus: oldStatus,
				newStatus: msg.Status,
				timestamp: time.Now(),
			})
			a.selectedTask.SetStatus(msg.Status)
			a.viewState = listView
			return a, a.updateTask(a.selectedTask)
		}
		return a, nil

	case components.StatusCancelledMsg:
		a.viewState = listView
		a.statusSelector = nil
		return a, nil

	case taskUpdatedMsg:
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		// Clear selections after update
		a.selectedTasks = make(map[string]*task.Task)
		// Reload tasks to reflect changes
		return a, a.loadTasks

	case undoMsg:
		return a, a.processUndo(msg.actions)

	case components.TaskCreatedMsg:
		a.viewState = listView
		a.taskForm = nil
		return a, a.createTask(msg.Task)

	case components.TaskFormCancelledMsg:
		a.viewState = listView
		a.taskForm = nil
		return a, nil

	case taskCreatedMsg:
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		// Reload tasks to show the new one
		return a, a.loadTasks

	case error:
		// TODO: Better error handling
		a.quitting = true
		return a, tea.Quit
	}

	var cmd tea.Cmd
	
	switch a.viewState {
	case listView:
		a.list, cmd = a.list.Update(msg)
	case detailView:
		if a.detail != nil {
			a.detail, cmd = a.detail.Update(msg)
		}
	case statusSelectView:
		if a.statusSelector != nil {
			a.statusSelector, cmd = a.statusSelector.Update(msg)
		}
	case createView:
		if a.taskForm != nil {
			a.taskForm, cmd = a.taskForm.Update(msg)
		}
	}
	
	return a, cmd
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	
	switch a.viewState {
	case detailView:
		if a.detail != nil {
			return a.detail.View()
		}
	case statusSelectView:
		if a.statusSelector != nil {
			title := lipgloss.NewStyle().Bold(true).Padding(1, 2).Render("Select Status")
			selector := lipgloss.NewStyle().Padding(1, 2).Render(a.statusSelector.View())
			help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(1, 2).
				Render("↑/k: up • ↓/j: down • enter: select • esc: cancel")
			return lipgloss.JoinVertical(lipgloss.Left, title, selector, help)
		}
	case createView:
		if a.taskForm != nil {
			return lipgloss.Place(a.width, a.height,
				lipgloss.Center, lipgloss.Center,
				a.taskForm.View())
		}
	case listView:
		listView := a.list.View()
		var statusInfo string
		
		// Show selected tasks count
		if len(a.selectedTasks) > 0 {
			statusInfo = fmt.Sprintf("%d tasks selected", len(a.selectedTasks))
		}
		
		// Show undo history count
		if len(a.undoHistory) > 0 {
			if statusInfo != "" {
				statusInfo += " | "
			}
			statusInfo += fmt.Sprintf("Undo available (%d)", len(a.undoHistory))
		}
		
		if statusInfo != "" {
			statusLine := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true).
				Padding(0, 2).
				Render(statusInfo)
			return lipgloss.JoinVertical(lipgloss.Left, listView, statusLine)
		}
		return listView
	}
	
	return a.list.View()
}

// Messages
type tasksLoadedMsg struct {
	tasks []*task.Task
}

type taskUpdatedMsg struct {
	err error
}

type undoMsg struct {
	actions []undoAction
}

type taskCreatedMsg struct {
	err error
}

// Commands
func (a *App) loadTasks() tea.Msg {
	tasks, err := a.repo.FindAll()
	if err != nil {
		return err
	}
	return tasksLoadedMsg{tasks: tasks}
}

func (a *App) updateTask(t *task.Task) tea.Cmd {
	return func() tea.Msg {
		err := a.repo.Update(t)
		return taskUpdatedMsg{err: err}
	}
}

func (a *App) updateTasks(tasks map[string]*task.Task, status task.Status) tea.Cmd {
	return func() tea.Msg {
		for _, t := range tasks {
			t.SetStatus(status)
			if err := a.repo.Update(t); err != nil {
				return taskUpdatedMsg{err: err}
			}
		}
		return taskUpdatedMsg{err: nil}
	}
}

func (a *App) undoLastAction() tea.Cmd {
	return func() tea.Msg {
		if len(a.undoHistory) == 0 {
			return nil
		}

		// Get the timestamp of the last action
		lastTimestamp := a.undoHistory[len(a.undoHistory)-1].timestamp
		
		// Collect all actions with the same timestamp (for bulk operations)
		var actionsToUndo []undoAction
		i := len(a.undoHistory) - 1
		for i >= 0 && a.undoHistory[i].timestamp.Equal(lastTimestamp) {
			actionsToUndo = append(actionsToUndo, a.undoHistory[i])
			i--
		}
		
		// Remove these actions from history
		a.undoHistory = a.undoHistory[:i+1]
		
		return undoMsg{actions: actionsToUndo}
	}
}

func (a *App) processUndo(actions []undoAction) tea.Cmd {
	return func() tea.Msg {
		for _, action := range actions {
			// Find the task and revert its status
			t, err := a.repo.FindByID(action.taskID)
			if err != nil {
				continue
			}
			t.SetStatus(action.oldStatus)
			if err := a.repo.Update(t); err != nil {
				return taskUpdatedMsg{err: err}
			}
		}
		return taskUpdatedMsg{err: nil}
	}
}

func (a *App) createTask(t *task.Task) tea.Cmd {
	return func() tea.Msg {
		_, err := a.repo.Create(t)
		return taskCreatedMsg{err: err}
	}
}

func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}