package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tkancf/mdtask/internal/task"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			PaddingTop(1).
			PaddingLeft(2)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2)

	contentStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1).
			PaddingLeft(2)
)

type keyMap struct {
	Back   key.Binding
	Edit   key.Binding
	Status key.Binding
	Help   key.Binding
	Quit   key.Binding
}

var keys = keyMap{
	Back: key.NewBinding(
		key.WithKeys("esc", "b"),
		key.WithHelp("esc/b", "back to list"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit task"),
	),
	Status: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "change status"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type DetailView struct {
	task     *task.Task
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

func NewDetailView(t *task.Task) *DetailView {
	return &DetailView{
		task: t,
	}
}

func (d *DetailView) Init() tea.Cmd {
	return nil
}

func (d *DetailView) Update(msg tea.Msg) (*DetailView, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

		if !d.ready {
			d.viewport = viewport.New(msg.Width, msg.Height-8)
			d.viewport.SetContent(d.renderContent())
			d.ready = true
		} else {
			d.viewport.Width = msg.Width
			d.viewport.Height = msg.Height - 8
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return d, GoBackCmd
		}
	}

	d.viewport, cmd = d.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return d, tea.Batch(cmds...)
}

func (d *DetailView) View() string {
	if !d.ready {
		return "\n  Loading..."
	}

	header := d.renderHeader()
	footer := d.renderFooter()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		d.viewport.View(),
		footer,
	)
}

func (d *DetailView) renderHeader() string {
	title := titleStyle.Render(d.task.Title)
	
	status := d.getTaskStatus()
	tags := strings.Join(d.getDisplayTags(), ", ")
	
	info := infoStyle.Render(fmt.Sprintf("ID: %s | Status: %s | Tags: %s", d.task.ID, status, tags))
	
	if deadline := d.task.GetDeadline(); deadline != nil {
		info += "\n" + infoStyle.Render(fmt.Sprintf("Deadline: %s", deadline.Format("2006-01-02")))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, title, info, "")
}

func (d *DetailView) renderContent() string {
	return contentStyle.Render(d.task.Content)
}

func (d *DetailView) renderFooter() string {
	help := helpStyle.Render("[esc] back • [e] edit • [s] status • [?] help • [q] quit")
	return help
}

func (d *DetailView) getTaskStatus() string {
	for _, tag := range d.task.Tags {
		switch tag {
		case "mdtask/status/TODO":
			return "TODO"
		case "mdtask/status/WIP":
			return "WIP"
		case "mdtask/status/WAIT":
			return "WAIT"
		case "mdtask/status/SCHE":
			return "SCHE"
		case "mdtask/status/DONE":
			return "DONE"
		}
	}
	return "Unknown"
}

func (d *DetailView) getDisplayTags() []string {
	var displayTags []string
	for _, tag := range d.task.Tags {
		if !strings.HasPrefix(tag, "mdtask/") {
			displayTags = append(displayTags, tag)
		}
	}
	return displayTags
}

// Command to go back to list view
func GoBackCmd() tea.Msg {
	return GoBackMsg{}
}

type GoBackMsg struct{}