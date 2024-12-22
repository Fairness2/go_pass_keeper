package card

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
	"strconv"
	"strings"
)

var (
	focusedButton    = style.ButtonFocusedStyle.Render("[ Сохранить ]")
	headerNewText    = style.HeaderStyle.Render("Новая карта")
	headerUpdateText = style.HeaderStyle.Render("Изменить карту")
	blurredButton    = style.ButtonBlurredStyle.Render("[ Сохранить ]")
)

const (
	ccnI = iota
	expI
	cvvI
	ownerI
	commentI
)

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	focusIndex int
	pService   *service.CRUDService[*payloads.CardWithComment, service.CardData]
	data       *payloads.CardWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service *service.CRUDService[*payloads.CardWithComment, service.CardData], data *payloads.CardWithComment) Form {
	number := components.NewTInput("**** **** **** ****", string(data.Number), true)
	number.CharLimit = 20
	number.Width = 30
	number.Prompt = ""
	number.Validate = ccnValidator
	date := components.NewTInput("MM/YY", string(data.Date), false)
	date.CharLimit = 5
	date.Width = 5
	date.Validate = expValidator
	cvv := components.NewTInput("XXX", string(data.CVV), false)
	cvv.CharLimit = 3
	cvv.Width = 3
	cvv.Validate = cvvValidator
	owner := components.NewTInput("OWNER", string(data.Owner), false)
	owner.CharLimit = 20
	owner.Width = 30
	owner.Prompt = ""
	comment := components.NewTArea("Comment", data.Comment, false)

	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			number,
			date,
			cvv,
			owner,
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
				return m.updateCard()
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

// updateCard обрабатывает ввод пользователя для создания или обновления записи карты, шифрует ее и возвращает соответствующую модель.
func (m Form) updateCard() (tea.Model, tea.Cmd) {
	m.data.Number = []byte(m.inputs[ccnI].Value())
	m.data.Date = []byte(m.inputs[expI].Value())
	m.data.CVV = []byte(m.inputs[cvvI].Value())
	m.data.Owner = []byte(m.inputs[ownerI].Value())
	m.data.Comment = m.inputs[commentI].Value()
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
	fmt.Fprintf(&b,
		`
 %s
 %s

 %s  %s
 %s  %s

 %s
 %s

%s
`,
		style.FocusedStyle.Width(30).Render("Номер карты"),
		m.inputs[ccnI].View(),
		style.FocusedStyle.Width(6).Render("EXP"),
		style.FocusedStyle.Width(6).Render("CVV"),
		m.inputs[expI].View(),
		m.inputs[cvvI].View(),
		style.FocusedStyle.Width(6).Render("Держатель"),
		m.inputs[ownerI].View(),
		m.inputs[commentI].View(),
	)
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

// Validator functions to ensure valid input
func ccnValidator(s string) error {
	// Credit Card Number should a string less than 20 digits
	// It should include 16 integers and 3 spaces
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}

	if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	// The last digit should be a number unless it is a multiple of 4 in which
	// case it should be a space
	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}

	// The remaining digits should be integers
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func expValidator(s string) error {
	// The 3 character should be a slash (/)
	// The rest should be numbers
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	// There should be only one slash and it should be in the 2nd index (3rd character)
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}

	return nil
}

func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	// Since the input will already ensure that the CVV is a string of length 3,
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}
