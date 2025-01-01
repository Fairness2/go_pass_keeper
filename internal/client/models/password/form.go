package password

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/service"
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
	focusIndex int
	pService   iFormService
	data       *payloads.PasswordWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service iFormService, data *payloads.PasswordWithComment) Form {
	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			components.NewTInput(loginPlaceholder, string(data.Username), true),
			components.NewTInput(passwordPlaceholder, string(data.Password.Password), false),
			components.NewTInput(domainPlaceholder, data.Domen, false),
			components.NewTArea(commentPlaceholder, data.Comment, false),
		},
		help:     help.New(),
		helpKeys: models.BaseFormHelp,
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
		case "ctrl+c", "esc":
			return m, tea.Quit
		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m.updatePassword()
			}
			// На комментарии нужно разрешать делать новую строку
			if s == "enter" && m.focusIndex == len(m.inputs)-1 {
				break
			}
			m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs), s)

			return m, models.GetCmds(m.inputs, m.focusIndex)
		}
	}
	// Handle character input and blinking
	cmd := models.UpdateInputs(msg, m.inputs)
	return m, cmd
}

// updatePassword обрабатывает ввод пользователя для создания или обновления записи пароля, шифрует ее и возвращает соответствующую модель.
func (m Form) updatePassword() (tea.Model, tea.Cmd) {
	m.data.Domen = m.inputs[domainI].Value()
	m.data.Username = []byte(m.inputs[loginI].Value())
	m.data.Password.Password = []byte(m.inputs[passwordI].Value())
	m.data.Comment = m.inputs[commentI].Value()
	var err error
	m.data, err = m.pService.EncryptItem(m.data)
	if err != nil {
		m.modelError = err
		return m, models.GetCmds(m.inputs, m.focusIndex)
	}
	if m.data.ID == "" {
		if err = m.pService.Create(m.data); err != nil {
			m.modelError = err
			return m, models.GetCmds(m.inputs, m.focusIndex)
		}
	} else {
		if err = m.pService.Update(m.data); err != nil {
			m.modelError = err
			return m, models.GetCmds(m.inputs, m.focusIndex)
		}
	}
	l := NewList(service.NewDefaultPasswordService())
	return l, l.Init()
}

func (m Form) View() string {
	return models.FormView(formCnf, m.data.ID, m.inputs, m.focusIndex, m.modelError, m.help.ShortHelpView(m.helpKeys))
}
