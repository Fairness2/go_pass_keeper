package password

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
)

const (
	loginI = iota
	passwordI
	domainI
	commentI
)

const (
	newTextHeader       = "Новый пароль"
	updateTextHeader    = "Изменить пароль"
	loginPlaceholder    = "Логин"
	passwordPlaceholder = "Пароль"
	domainPlaceholder   = "Домен"
	commentPlaceholder  = "Comment"
)

var (
	focusedButton    = style.ButtonFocusedStyle.Render(models.SaveText)
	headerNewText    = style.HeaderStyle.Render(newTextHeader)
	headerUpdateText = style.HeaderStyle.Render(updateTextHeader)
	blurredButton    = style.ButtonBlurredStyle.Render(models.SaveText)

	formCnf = &models.FormViewConfig{
		HeaderNewText:    headerNewText,
		HeaderUpdateText: headerUpdateText,
		BlurredButton:    blurredButton,
		FocusedButton:    focusedButton,
	}
)

type iFormService interface {
	EncryptItem(body *payloads.PasswordWithComment) (*payloads.PasswordWithComment, error)
	Create(body *payloads.PasswordWithComment) error
	Update(body *payloads.PasswordWithComment) error
}

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	models.Backable
	focusIndex int
	pService   iFormService
	data       *payloads.PasswordWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service iFormService, data *payloads.PasswordWithComment, lastModel tea.Model) Form {
	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			components.NewTInput(loginPlaceholder, string(data.Username), true),
			components.NewTPass(passwordPlaceholder, string(data.Password.Password), false),
			components.NewTInput(domainPlaceholder, data.Domen, false),
			components.NewTArea(commentPlaceholder, data.Comment, false),
		},
		help:     help.New(),
		helpKeys: models.BaseFormHelp,
		Backable: models.NewBackable(lastModel),
	}

	return m
}

// Init инициализирует состояние формы и возвращает команду начать мигать для активного поля ввода.
func (m Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update обрабатывает входящие сообщения для обновления состояния формы, обработки входных данных и выполнения операций на основе действий пользователя.
func (m Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.Back()
		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			return m.navigationMessage(msg)
		}
	}
	// Handle character input and blinking
	cmd := models.UpdateInputs(msg, m.inputs)
	return m, cmd
}

// navigationMessage обрабатывает события нажатия клавиш для навигации по полям ввода в форме и соответствующим образом запускает обновления фокуса полей или действия.
func (m Form) navigationMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()
	if s == "enter" && m.focusIndex == len(m.inputs) {
		return m.updatePassword()
	}
	// На комментарии нужно разрешать делать новую строку
	if s == "enter" && m.focusIndex == len(m.inputs)-1 {
		return m, models.UpdateInputs(msg, m.inputs)
	}
	m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs), s)

	return m, models.GetCmds(m.inputs, m.focusIndex)
}

// updatePassword обрабатывает ввод пользователя для создания или обновления записи пароля, шифрует ее и возвращает соответствующую модель.
func (m Form) updatePassword() (tea.Model, tea.Cmd) {
	m.data.Domen = m.inputs[domainI].Value()
	m.data.Username = []byte(m.inputs[loginI].Value())
	m.data.Password.Password = []byte(m.inputs[passwordI].Value())
	m.data.Comment = m.inputs[commentI].Value()
	var err error
	m.data, err = m.pService.EncryptItem(m.data)
	if err == nil {
		if m.data.ID == "" {
			err = m.pService.Create(m.data)
		} else {
			err = m.pService.Update(m.data)
		}
	}
	if err != nil {
		m.modelError = err
		return m, models.GetCmds(m.inputs, m.focusIndex)
	}
	return m.Back()
}

func (m Form) View() string {
	return models.FormView(formCnf, m.data.ID, m.inputs, m.focusIndex, m.modelError, m.help.ShortHelpView(m.helpKeys))
}
