package login

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/config"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/models/menu"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"strings"
)

var (
	buttonLogin           = "[ Вход ]"
	buttonRegister        = "[ Регистрация ]"
	headerText            = style.HeaderStyle.Render("Вход или регистрация")
	focusedLoginButton    = style.ButtonFocusedStyle.Render(buttonLogin)
	blurredLoginButton    = style.ButtonBlurredStyle.Render(buttonLogin)
	focusedRegisterButton = style.ButtonFocusedStyle.Render(buttonRegister)
	blurredRegisterButton = style.ButtonBlurredStyle.Render(buttonRegister)
)

// Model представляет собой основную структуру для управления состоянием пользовательского интерфейса входа в систему и его взаимодействия с LoginService.
type Model struct {
	focusIndex int
	inputs     []*components.TInput
	service    *service.LoginService
	modelError error
	help       help.Model
	helpKeys   []key.Binding
	showInfo   bool
}

// InitialModel инициализирует новую модель с предопределенными полями ввода имени и пароля и назначает LoginService.
func InitialModel(service *service.LoginService) Model {
	m := Model{
		inputs: []*components.TInput{
			components.NewTInput("Логин", "", true),
			components.NewTPass("Пароль", "", false),
		},
		service: service,
		help:    help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("ctrl+c, esc", "Выход"), key.WithKeys("ctrl+c", "esc")),
			key.NewBinding(key.WithHelp("tab, shift+tab, up, down", "Переход по форме"), key.WithKeys("tab", "shift+tab", "up", "down")),
			key.NewBinding(key.WithHelp("enter", "Принять"), key.WithKeys("enter")),
			key.NewBinding(key.WithHelp("f1", "Показать/скрыть информацию о клиенте"), key.WithKeys("f1")),
		},
	}
	return m
}

// Init инициализирует Model, возвращая команду мигающего курсора для ввода текста.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update обрабатывает входящие сообщения и соответствующим образом обновляет состояние Model, возвращая обновленную Model и tea.Cmd.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m.authorize(false)
			}
			if s == "enter" && m.focusIndex == len(m.inputs)+1 {
				return m.authorize(true)
			}
			m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs)+1, s)
			return m, m.getCmds()
		case "f1":
			m.showInfo = !m.showInfo
			return m, nil
		}
	}
	// Handle character input and blinking
	cmd := m.updateInputs(msg)
	return m, cmd
}

// getCmds возвращает пакет команд для обновления состояния фокуса входных данных на основе текущего индекса фокуса в модели.
func (m Model) getCmds() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
	}
	return tea.Batch(cmds...)
}

// updateInputs обновляет состояние каждого поля ввода на основе предоставленного сообщения и собирает их команды.
func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// View генерирует визуальное представление Model, включая поля ввода, кнопку и сообщение об ошибке, если применимо.
func (m Model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", headerText)
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
	}
	if m.modelError != nil {
		fmt.Fprintf(&b, "\n%s\n", style.ErrorStyle.Render(m.modelError.Error()))
	}

	if m.showInfo {
		fmt.Fprintf(&b, "\n%s\n%s\n%s\n%s\n%s\n%s\n",
			style.BlurredStyle.Render("BuildDate: "+config.BuildDate),
			style.BlurredStyle.Render("BuildCommit: "+config.BuildCommit),
			style.BlurredStyle.Render("BuildVersion: "+config.BuildVersion),
			style.BlurredStyle.Render("BuildOS: "+config.BuildOS),
			style.BlurredStyle.Render("ServerAddress: "+config.ServerAddress),
			style.BlurredStyle.Render("LogLevel: "+config.LogLevel),
		)
	}
	loginButton := blurredLoginButton
	if m.focusIndex == len(m.inputs) {
		loginButton = focusedLoginButton
	}
	registerButton := blurredRegisterButton
	if m.focusIndex == len(m.inputs)+1 {
		registerButton = focusedRegisterButton
	}
	fmt.Fprintf(&b, "\n%s     %s\n\n", loginButton, registerButton)

	b.WriteString(m.help.ShortHelpView(m.helpKeys))

	return b.String()
}

// login пытается авторизовать пользователя, используя предоставленные логин и пароль, и в случае неудачи возвращает ошибку.
func (m Model) login(isRegistration bool) error {
	login := m.inputs[0].Value()
	password := m.inputs[1].Value()
	return m.service.Login(login, password, isRegistration)
}

// authorize пытается авторизовать в систему пользователя и в случае успеха переходит к новой модели меню, в противном случае возвращает ошибки.
func (m Model) authorize(isRegistration bool) (tea.Model, tea.Cmd) {
	if err := m.login(isRegistration); err != nil {
		m.modelError = err
		return m, m.getCmds()
	}
	newM := menu.NewModel()
	return newM, newM.Init()
}
