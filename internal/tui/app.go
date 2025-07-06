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
)

const (
	listHeight = 20
	listWidth  = 80
)

type App struct {
	repo     repository.Repository
	list     list.Model
	tasks    []*task.Task
	width    int
	height   int
	quitting bool
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
		a.list.SetWidth(msg.Width)
		a.list.SetHeight(msg.Height - 2)
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			a.quitting = true
			return a, tea.Quit
		}

	case tasksLoadedMsg:
		a.tasks = msg.tasks
		items := make([]list.Item, len(a.tasks))
		for i, t := range a.tasks {
			items[i] = taskItem{task: t}
		}
		a.list.SetItems(items)
		return a, nil

	case error:
		// TODO: Better error handling
		a.quitting = true
		return a, tea.Quit
	}

	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	return a, cmd
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	return a.list.View()
}

// Messages
type tasksLoadedMsg struct {
	tasks []*task.Task
}

// Commands
func (a *App) loadTasks() tea.Msg {
	tasks, err := a.repo.FindAll()
	if err != nil {
		return err
	}
	return tasksLoadedMsg{tasks: tasks}
}

func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}