package components

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tkancf/mdtask/internal/task"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("99")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Cancel key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "select"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "cancel"),
	),
}

type StatusSelector struct {
	statuses []task.Status
	cursor   int
	selected task.Status
}

func NewStatusSelector(current task.Status) *StatusSelector {
	statuses := []task.Status{
		task.StatusTODO,
		task.StatusWIP,
		task.StatusWAIT,
		task.StatusSCHE,
		task.StatusDONE,
	}

	cursor := 0
	for i, s := range statuses {
		if s == current {
			cursor = i
			break
		}
	}

	return &StatusSelector{
		statuses: statuses,
		cursor:   cursor,
		selected: current,
	}
}

func (s *StatusSelector) Init() tea.Cmd {
	return nil
}

func (s *StatusSelector) Update(msg tea.Msg) (*StatusSelector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if s.cursor > 0 {
				s.cursor--
			}
		case key.Matches(msg, keys.Down):
			if s.cursor < len(s.statuses)-1 {
				s.cursor++
			}
		case key.Matches(msg, keys.Select):
			s.selected = s.statuses[s.cursor]
			return s, StatusSelectedCmd(s.selected)
		case key.Matches(msg, keys.Cancel):
			return s, StatusCancelledCmd
		}
	}
	return s, nil
}

func (s *StatusSelector) View() string {
	var items string
	for i, status := range s.statuses {
		label := string(status)
		if i == s.cursor {
			items += selectedStyle.Render(label) + "\n"
		} else {
			items += normalStyle.Render(label) + "\n"
		}
	}
	return items
}

// Messages
type StatusSelectedMsg struct {
	Status task.Status
}

type StatusCancelledMsg struct{}

// Commands
func StatusSelectedCmd(status task.Status) tea.Cmd {
	return func() tea.Msg {
		return StatusSelectedMsg{Status: status}
	}
}

func StatusCancelledCmd() tea.Msg {
	return StatusCancelledMsg{}
}