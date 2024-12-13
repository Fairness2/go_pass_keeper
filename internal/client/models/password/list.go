package password

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"passkeeper/internal/client/service"
	"passkeeper/internal/payloads"
	"strings"
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fc0303"))
)

type PassData struct {
	payloads.PasswordWithComment
}

func (i PassData) Title() string       { return i.Domen }
func (i PassData) Description() string { return i.Comment }
func (i PassData) FilterValue() string { return i.Domen }

type List struct {
	list     list.Model
	pService *service.PasswordService
	refErr   error
}

func (m List) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m List) View() string {
	var b strings.Builder
	if m.refErr != nil {
		fmt.Fprintf(&b, "\n\n%s\n\n", errorStyle.Render(m.refErr.Error()))
	}
	b.WriteString(docStyle.Render(m.list.View()))
	return b.String()
}

func NewList(passwordService *service.PasswordService) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: passwordService,
	}
	m.list.Title = "Список паролей"

	if err := m.refresh(); err != nil {
		m.refErr = err
	}
	return m
}

func (l *List) refresh() error {
	passwords, err := l.pService.GetPasswords()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(passwords))
	for i, p := range passwords {
		pl[i] = PassData{p}
	}
	l.list.SetItems(pl)
	return nil
}
