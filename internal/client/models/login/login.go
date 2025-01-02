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
	"passkeeper/internal/client/style"
	"strings"
)

const (
	loginI = iota
	passwordI
)

const (
	buttonLogin    = "[ Вход ]"
	buttonRegister = "[ Регистрация ]"
	header         = "Вход или регистрация"
	loginText      = "Логин"
	passwordText   = "Пароль"
	infoText       = "Показать/скрыть информацию о клиенте"
)

var (
	headerText            = style.HeaderStyle.Render(header)
	focusedLoginButton    = style.ButtonFocusedStyle.Render(buttonLogin)
	blurredLoginButton    = style.ButtonBlurredStyle.Render(buttonLogin)
	focusedRegisterButton = style.ButtonFocusedStyle.Render(buttonRegister)
	blurredRegisterButton = style.ButtonBlurredStyle.Render(buttonRegister)
)

// processService определяет интерфейс для обработки аутентификации пользователя с поддержкой процесса входа в систему и регистрации.
type processService interface {
	Login(username, password string, isRegistration bool) error
}

// Model представляет собой основную структуру для управления состоянием пользовательского интерфейса входа в систему и его взаимодействия с LoginService.
type Model struct {
	focusIndex int
	inputs     []components.BlinkInput
	service    processService
	modelError error
	help       help.Model
	helpKeys   []key.Binding
	showInfo   bool
}

// InitialModel инициализирует новую модель с предопределенными полями ввода имени и пароля и назначает LoginService.
func InitialModel(service processService) Model {
	helps := make([]key.Binding, 0, 4)
	helps = append(helps, models.BaseFormHelp...)
	helps = append(helps, key.NewBinding(key.WithHelp("f1", infoText), key.WithKeys("f1")))
	m := Model{
		inputs: []components.BlinkInput{
			components.NewTInput(loginText, "", true),
			components.NewTPass(passwordText, "", false),
		},
		service:  service,
		help:     help.New(),
		helpKeys: helps,
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
			return m.navigationMessage(msg)
		case "f1":
			m.showInfo = !m.showInfo
			return m, nil
		}
	}
	// Handle character input and blinking
	cmd := models.UpdateInputs(msg, m.inputs)
	return m, cmd
}

// navigationMessage обрабатывает события нажатия клавиш для навигации по полям ввода в форме и соответствующим образом запускает обновления фокуса полей или действия.
func (m Model) navigationMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "enter" && m.focusIndex == len(m.inputs) {
		return m.authorize(false)
	}
	if s == "enter" && m.focusIndex == len(m.inputs)+1 {
		return m.authorize(true)
	}
	m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs)+1, s)
	return m, models.GetCmds(m.inputs, m.focusIndex)
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
	login := m.inputs[loginI].Value()
	password := m.inputs[passwordI].Value()
	return m.service.Login(login, password, isRegistration)
}

// authorize пытается авторизовать в систему пользователя и в случае успеха переходит к новой модели меню, в противном случае возвращает ошибки.
func (m Model) authorize(isRegistration bool) (tea.Model, tea.Cmd) {
	if err := m.login(isRegistration); err != nil {
		m.modelError = err
		return m, models.GetCmds(m.inputs, m.focusIndex)
	}
	newM := menu.NewModel()
	return newM, newM.Init()
}
