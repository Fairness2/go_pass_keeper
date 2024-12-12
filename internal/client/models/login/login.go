package login

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"passkeeper/internal/client/models/menu"
	"passkeeper/internal/client/service"
	"strings"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#fc0303"))

	focusedButton = focusedStyle.Render("[ Вход ]")
	headerText    = focusedStyle.Render("Вход в систему")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Вход"))
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	service    *service.LoginService
	lgnErr     error
}

func InitialModel(service *service.LoginService) Model {
	m := Model{
		inputs:  make([]textinput.Model, 2),
		service: service,
	}
	lgn := textinput.New()
	lgn.Cursor.Style = cursorStyle
	lgn.Placeholder = "Логин"
	lgn.Focus()
	lgn.PromptStyle = focusedStyle
	lgn.TextStyle = focusedStyle
	m.inputs[0] = lgn

	pass := textinput.New()
	pass.Cursor.Style = cursorStyle
	pass.Placeholder = "Пароль"
	pass.PromptStyle = focusedStyle
	pass.TextStyle = focusedStyle
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '•'
	m.inputs[1] = pass

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				if err := m.login(); err != nil {
					m.lgnErr = err
					return m, m.getCmds()
				}
				newM := menu.NewModel()
				return newM, newM.Init()
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			return m, m.getCmds()
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m Model) getCmds() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}
	return tea.Batch(cmds...)
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n\n", headerText)

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	if m.lgnErr != nil {
		fmt.Fprintf(&b, "\n\n%s\n\n", errorStyle.Render(m.lgnErr.Error()))
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func (m *Model) login() error {
	login := m.inputs[0].Value()
	password := m.inputs[1].Value()
	return m.service.Login(login, password)
}
