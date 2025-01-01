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

const (
	ccnI = iota
	expI
	cvvI
	ownerI
	commentI
)

const (
	newTextHeader      = "Новая карта"
	updateTextHeader   = "Изменить карту"
	numberPlaceholder  = "**** **** **** ****"
	datePlaceholder    = "MM/YY"
	ownerPlaceholder   = "OWNER"
	commentPlaceholder = "Comment"
	cvvPlaceholder     = "XXX"
	numberFieldName    = "Номер карты"
	dateFiledName      = "EXP"
	cvvFiledName       = "CVV"
	ownerFieldName     = "Держатель"
	formTemplate       = `%s

 %s
 %s

 %s  %s
 %s  %s

 %s
 %s

%s

%s
%s

%s
`
)

var (
	focusedButton    = style.ButtonFocusedStyle.Render(models.SaveText)
	headerNewText    = style.HeaderStyle.Render(newTextHeader)
	headerUpdateText = style.HeaderStyle.Render(updateTextHeader)
	blurredButton    = style.ButtonBlurredStyle.Render(models.SaveText)
)

type iFormService interface {
	EncryptItem(body *payloads.CardWithComment) (*payloads.CardWithComment, error)
	Create(body *payloads.CardWithComment) error
	Update(body *payloads.CardWithComment) error
}

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	focusIndex int
	pService   iFormService
	data       *payloads.CardWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service iFormService, data *payloads.CardWithComment) Form {
	number := components.NewTInput(numberPlaceholder, string(data.Number), true)
	number.CharLimit = 20
	number.Width = 30
	number.Prompt = ""
	number.Validate = ccnValidator
	date := components.NewTInput(datePlaceholder, string(data.Date), false)
	date.CharLimit = 5
	date.Width = 5
	date.Validate = expValidator
	cvv := components.NewTInput(cvvPlaceholder, string(data.CVV), false)
	cvv.CharLimit = 3
	cvv.Width = 3
	cvv.Validate = cvvValidator
	owner := components.NewTInput(ownerPlaceholder, string(data.Owner), false)
	owner.CharLimit = 20
	owner.Width = 30
	owner.Prompt = ""
	comment := components.NewTArea(commentPlaceholder, data.Comment, false)

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
				return m.updateCard()
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
	l := NewList(service.NewDefaultCardService())
	return l, l.Init()
}

// View отображает форму на основе ее текущего состояния, включая входные данные, кнопки и ошибки, и возвращает визуализированную строку.
func (m Form) View() string {
	// Выбираем текст заголовка
	h := headerNewText
	if m.data.ID != "" {
		h = headerUpdateText
	}
	// Составляем текст ошибки
	var errStr string
	if m.modelError != nil {
		errStr = style.ErrorStyle.Render(m.modelError.Error())
	}

	// Составляем текст кнопки
	button := blurredButton
	if m.focusIndex == len(m.inputs) {
		button = focusedButton
	}

	return fmt.Sprintf(formTemplate,
		h,
		style.FocusedStyle.Width(30).Render(numberFieldName),
		m.inputs[ccnI].View(),
		style.FocusedStyle.Width(6).Render(dateFiledName),
		style.FocusedStyle.Width(6).Render(cvvFiledName),
		m.inputs[expI].View(),
		m.inputs[cvvI].View(),
		style.FocusedStyle.Width(10).Render(ownerFieldName),
		m.inputs[ownerI].View(),
		m.inputs[commentI].View(),
		errStr,
		button,
		m.help.ShortHelpView(m.helpKeys))
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

// expValidator проверяет формат даты истечения срока действия (ММ/ГГ), обеспечивая одинарную косую черту в третьей позиции и числовые значения.
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

// cvvValidator проверяет, является ли входная строка числовым значением, состоящим ровно из 3 цифр.
func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	// Since the input will already ensure that the CVV is a string of length 3,
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}
