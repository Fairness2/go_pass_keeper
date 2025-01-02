package text

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
	textI = iota
	commentI
)

const (
	newTextHeader        = "Новый текст"
	updateTextHeader     = "Изменить текст"
	textInputPlaceholder = "Text"
	commentPlaceholder   = "Comment"
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
	EncryptItem(body *payloads.TextWithComment) (*payloads.TextWithComment, error)
	Create(body *payloads.TextWithComment) error
	Update(body *payloads.TextWithComment) error
}

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	focusIndex int
	pService   iFormService
	data       *payloads.TextWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service iFormService, data *payloads.TextWithComment) Form {
	text := components.NewTArea(textInputPlaceholder, string(data.TextData), true)
	comment := components.NewTArea(commentPlaceholder, data.Comment, false)

	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			text,
			comment,
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
		return m.updateText()
	}
	// На комментарии нужно разрешать делать новую строку
	if s == "enter" && m.focusIndex == len(m.inputs)-1 {
		return m, models.UpdateInputs(msg, m.inputs)
	}
	m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs), s)

	return m, models.GetCmds(m.inputs, m.focusIndex)
}

// updateText обрабатывает ввод пользователя для создания или обновления записи текста, шифрует ее и возвращает соответствующую модель.
func (m Form) updateText() (tea.Model, tea.Cmd) {
	m.data.TextData = []byte(m.inputs[textI].Value())
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
	l := NewList(service.NewDefaultTextService())
	return l, l.Init()
}

// View отображает форму на основе ее текущего состояния, включая входные данные, кнопки и ошибки, и возвращает визуализированную строку.
func (m Form) View() string {
	return models.FormView(formCnf, m.data.ID, m.inputs, m.focusIndex, m.modelError, m.help.ShortHelpView(m.helpKeys))
}
