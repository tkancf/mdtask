package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tkancf/mdtask/internal/task"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	labelStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(12)
	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
)

type formKeyMap struct {
	Next   key.Binding
	Prev   key.Binding
	Submit key.Binding
	Cancel key.Binding
}

var formKeys = formKeyMap{
	Next: key.NewBinding(
		key.WithKeys("tab", "down"),
		key.WithHelp("tab/↓", "next field"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab", "up"),
		key.WithHelp("shift+tab/↑", "prev field"),
	),
	Submit: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save task"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

type TaskForm struct {
	focusIndex int
	inputs     []textinput.Model
	width      int
	height     int
}

func NewTaskForm() *TaskForm {
	m := &TaskForm{
		inputs: make([]textinput.Model, 3),
	}

	// Title input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Task title (required)"
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 100
	m.inputs[0].Width = 50
	m.inputs[0].PromptStyle = focusedStyle
	m.inputs[0].TextStyle = focusedStyle

	// Description input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Brief description (optional)"
	m.inputs[1].CharLimit = 200
	m.inputs[1].Width = 50

	// Tags input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Tags separated by comma (optional)"
	m.inputs[2].CharLimit = 100
	m.inputs[2].Width = 50

	return m
}

func (m *TaskForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TaskForm) Update(msg tea.Msg) (*TaskForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, formKeys.Submit):
			if m.inputs[0].Value() != "" {
				return m, m.createTask()
			}
		case key.Matches(msg, formKeys.Cancel):
			return m, TaskFormCancelledCmd
		case key.Matches(msg, formKeys.Next):
			m.focusIndex++
			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			}
			return m, m.updateFocus()
		case key.Matches(msg, formKeys.Prev):
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			return m, m.updateFocus()
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *TaskForm) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = noStyle
			m.inputs[i].TextStyle = noStyle
		}
	}
	return tea.Batch(cmds...)
}

func (m *TaskForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *TaskForm) View() string {
	var fields []string

	// Title field
	fields = append(fields, lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Title:"),
		m.inputs[0].View(),
	))

	// Description field
	fields = append(fields, lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Description:"),
		m.inputs[1].View(),
	))

	// Tags field
	fields = append(fields, lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render("Tags:"),
		m.inputs[2].View(),
	))

	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Render("tab/↓: next • shift+tab/↑: prev • ctrl+s: save • esc: cancel")

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		"Create New Task",
		"",
		strings.Join(fields, "\n\n"),
		help,
	)

	return formStyle.Render(form)
}

func (m *TaskForm) createTask() tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		taskID := fmt.Sprintf("task/%s", now.Format("20060102150405"))

		newTask := &task.Task{
			ID:          taskID,
			Title:       m.inputs[0].Value(),
			Description: m.inputs[1].Value(),
			Tags:        []string{"mdtask", "mdtask/status/TODO"},
			Created:     now,
			Updated:     now,
			Content:     fmt.Sprintf("# %s\n\n%s", m.inputs[0].Value(), m.inputs[1].Value()),
		}

		// Add custom tags if provided
		if tagsInput := strings.TrimSpace(m.inputs[2].Value()); tagsInput != "" {
			tags := strings.Split(tagsInput, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" && tag != "mdtask" && !strings.HasPrefix(tag, "mdtask/") {
					newTask.Tags = append(newTask.Tags, tag)
				}
			}
		}

		return TaskCreatedMsg{Task: newTask}
	}
}

// Messages
type TaskCreatedMsg struct {
	Task *task.Task
}

type TaskFormCancelledMsg struct{}

// Commands
func TaskFormCancelledCmd() tea.Msg {
	return TaskFormCancelledMsg{}
}