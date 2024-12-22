package password

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
	"strings"
)

var (
	focusedButton    = style.ButtonFocusedStyle.Render("[ Сохранить ]")
	headerNewText    = style.HeaderStyle.Render("Новый пароль")
	headerUpdateText = style.HeaderStyle.Render("Изменить пароль")
	blurredButton    = style.ButtonBlurredStyle.Render("[ Сохранить ]")
)

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	focusIndex int
	pService   *service.CRUDService[*payloads.PasswordWithComment, service.PassData]
	data       *payloads.PasswordWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service *service.CRUDService[*payloads.PasswordWithComment, service.PassData], data *payloads.PasswordWithComment) Form {
	lgn := components.NewTInput("Логин", string(data.Username), true)
	domen := components.NewTInput("Домен", data.Domen, false)
	pass := components.NewTInput("Пароль", string(data.Password.Password), false)
	comment := components.NewTArea("Comment", data.Comment, false)

	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			lgn,
			pass,
			domen,
			comment,
		},
		help: help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("ctrl+c, esc", "Выход"), key.WithKeys("ctrl+c", "esc")),
			key.NewBinding(key.WithHelp("tab, shift+tab, up, down", "Переход по форме"), key.WithKeys("tab", "shift+tab", "up", "down")),
			key.NewBinding(key.WithHelp("enter", "Принять"), key.WithKeys("enter")),
		},
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

			return m, m.getCmds()
		}
	}
	// Handle character input and blinking
	cmd := m.updateInputs(msg)
	return m, cmd
}

// updatePassword обрабатывает ввод пользователя для создания или обновления записи пароля, шифрует ее и возвращает соответствующую модель.
func (m Form) updatePassword() (tea.Model, tea.Cmd) {
	m.data.Domen = m.inputs[2].Value()
	m.data.Username = []byte(m.inputs[0].Value())
	m.data.Password.Password = []byte(m.inputs[1].Value())
	m.data.Comment = m.inputs[3].Value()
	var err error
	m.data, err = m.pService.EncryptItem(m.data)
	if err != nil {
		m.modelError = err
		return m, m.getCmds()
	}
	if m.data.ID == 0 {
		if err = m.pService.Create(m.data); err != nil {
			m.modelError = err
			return m, m.getCmds()
		}
	} else {
		if err = m.pService.Update(m.data); err != nil {
			m.modelError = err
			return m, m.getCmds()
		}
	}
	l := NewList(m.pService)
	return l, l.Init()
}

// getCmds генерирует пакетную команду для обновления состояния фокуса входных данных формы на основе текущего индекса фокуса.
func (m Form) getCmds() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i, input := range m.inputs {
		if m.focusIndex == i {
			cmds[i] = input.Focus()
		} else {
			input.Blur()
		}
	}
	return tea.Batch(cmds...)
}

// updateInputs обновляет все входные компоненты в форме на основе предоставленного сообщения и возвращает пакетную команду для обновлений.
func (m *Form) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i, input := range m.inputs {
		switch r := input.(type) {
		case *components.TInput:
			m.inputs[i], cmds[i] = r.Update(msg)
		case *components.TArea:
			m.inputs[i], cmds[i] = r.Update(msg)
		}
	}
	return tea.Batch(cmds...)
}

// View отображает форму на основе ее текущего состояния, включая входные данные, кнопки и ошибки, и возвращает визуализированную строку.
func (m Form) View() string {
	var b strings.Builder
	if m.data.ID != 0 {
		fmt.Fprintf(&b, "%s\n\n", headerUpdateText)
	} else {
		fmt.Fprintf(&b, "%s\n\n", headerNewText)
	}
	for _, input := range m.inputs {
		b.WriteString(input.View())
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.modelError != nil {
		fmt.Fprintf(&b, "%s\n\n", style.ErrorStyle.Render(m.modelError.Error()))
	}
	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "%s\n\n", *button)
	b.WriteString(m.help.ShortHelpView(m.helpKeys))

	return b.String()
}
