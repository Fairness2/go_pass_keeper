package menu

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models/card"
	"passkeeper/internal/client/models/file"
	"passkeeper/internal/client/models/password"
	"passkeeper/internal/client/models/text"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"strings"
)

var (
	headerText = style.HeaderStyle.Render("Выберите режим")
)

// Model представляет собой основную структуру, содержащую компоненты пользовательского интерфейса и привязки клавиш для навигации и взаимодействия.
type Model struct {
	cb       *components.Checkbox
	help     help.Model
	helpKeys []key.Binding
}

// NewModel инициализирует и возвращает новый экземпляр Model с настройками по умолчанию и привязками клавиш для навигации.
func NewModel() Model {
	return Model{
		cb: components.NewCheckbox(0,
			"Пары логин/пароль",
			"Произвольные текстовые данные",
			"Произвольные бинарные данные",
			"Данные банковских карт"),
		help: help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("ctrl+c, esc", "Выход"), key.WithKeys("ctrl+c", "esc")),
			key.NewBinding(key.WithHelp("j/k, up/down", "Выбор"), key.WithKeys("j", "down", "k", "up")),
			key.NewBinding(key.WithHelp("enter", "Принять"), key.WithKeys("enter")),
		},
	}
}

// Init инициализирует модель и возвращает команду, которая будет выполнена при запуске программы.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update обрабатывает ввод пользователя и соответствующим образом обновляет состояние модели. Он может вернуть команду для выполнения.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.nextView()
		}
	}

	m.cb, _ = m.cb.Update(msg)

	return m, nil
}

// View отображает полное представление модели, включая заголовок, компонент флажка и краткую справочную информацию.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(headerText)
	b.WriteString("\n\n")
	b.WriteString(m.cb.View())
	b.WriteString("\n\n")
	b.WriteString(m.help.ShortHelpView(m.helpKeys))

	return b.String()
}

// nextView определяет следующее представление на основе выбора флажка и инициализирует его;
// возвращает новую модель и команду.
func (m Model) nextView() (tea.Model, tea.Cmd) {
	switch m.cb.GetChoice() {
	case 0:
		l := password.NewList(service.NewDefaultPasswordService())
		return l, l.Init()
	case 1:
		l := text.NewList(service.NewDefaultTextService())
		return l, l.Init()
	case 2:
		l := file.NewList(service.NewDefaultFileService())
		return l, l.Init()
	case 3:
		l := card.NewList(service.NewDefaultCardService())
		return l, l.Init()
	default:
		return m, nil
	}
}
