package menu

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

const (
	dotChar = " • "
)

// General stuff for styling the view
var (
	keywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ticksStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("79"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
)

type (
	tickMsg  struct{}
	frameMsg struct{}
)

func NewModel() Model {
	return Model{0, false}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type Model struct {
	Choice int
	Chosen bool
}

func (m Model) Init() tea.Cmd {
	return tick()
}

// Main update function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.Choice++
			if m.Choice > 3 {
				m.Choice = 3
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "enter":
			m.Chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// The main view, which just calls the appropriate sub-view
func (m Model) View() string {
	var s string
	s = choicesView(m)
	return mainStyle.Render("\n" + s + "\n\n")
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m Model) string {
	c := m.Choice

	tpl := "Выберите режим?\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("j/k, up/down: выбор") + dotStyle +
		subtleStyle.Render("enter: выбрать") + dotStyle +
		subtleStyle.Render("q, esc: выйти")

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Пары логин/пароль", c == 0),
		checkbox("Произвольные текстовые данные", c == 1),
		checkbox("Произвольные бинарные данные", c == 2),
		checkbox("Данные банковских карт", c == 3),
	)

	return fmt.Sprintf(tpl, choices)
}

func checkbox(label string, checked bool) string {
	if checked {
		return checkboxStyle.Render("[x] " + label)
	}
	return fmt.Sprintf("[ ] %s", label)
}
