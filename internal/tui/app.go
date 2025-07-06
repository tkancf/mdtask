package tui

import (
	"fmt"
	"io"

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
)

type App struct {
	repo           repository.Repository
	list           list.Model
	tasks          []*task.Task
	detail         *views.DetailView
	statusSelector *components.StatusSelector
	selectedTask   *task.Task
	viewState      viewState
	width          int
	height         int
	quitting       bool
	err            error
}

// taskItem implements list.Item interface
type taskItem struct {
	task *task.Task
}

func (i taskItem) FilterValue() string { return i.task.Title }
func (i taskItem) Title() string       { return i.task.Title }
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
				key.WithKeys("enter", "v"),
				key.WithHelp("enter/v", "view task"),
			),
			key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "change status"),
			),
		}
	}

	return &App{
		repo: repo,
		list: l,
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
			case "enter", "v":
				// Open task detail view
				if i, ok := a.list.SelectedItem().(taskItem); ok {
					a.detail = views.NewDetailView(i.task)
					a.selectedTask = i.task
					a.viewState = detailView
					return a, a.detail.Init()
				}
			case "s":
				// Open status selector
				if i, ok := a.list.SelectedItem().(taskItem); ok {
					a.selectedTask = i.task
					a.statusSelector = components.NewStatusSelector(a.selectedTask.GetStatus())
					a.viewState = statusSelectView
					return a, a.statusSelector.Init()
				}
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
		if a.selectedTask != nil {
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
		// Reload tasks to reflect changes
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
	case listView:
		return a.list.View()
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

func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}