package file

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
	"strings"
)

const (
	filePathI = iota
	nameI
	commentI
)

const (
	newTextHeader    = "Новый файл"
	updateTextHeader = "Изменить файл"
)

var (
	focusedButton      = style.ButtonFocusedStyle.Render(models.SaveText)
	headerNewText      = style.HeaderStyle.Render(newTextHeader)
	headerUpdateText   = style.HeaderStyle.Render(updateTextHeader)
	blurredButton      = style.ButtonBlurredStyle.Render(models.SaveText)
	pathPlaceholder    = "Путь к файлу"
	namePlaceholder    = "Название"
	commentPlaceholder = "Comment"
)

type iFormService interface {
	EncryptItem(body *payloads.FileWithComment) (*payloads.FileWithComment, error)
	EncryptFile(filePath string) (string, error)
	CreateFile(body *payloads.FileWithComment, filePath string) error
	Update(body *payloads.FileWithComment) error
}

// Form представляет собой структуру для управления формами пользовательского ввода, включая управление фокусом и проверку ввода.
type Form struct {
	models.Backable
	focusIndex int
	pService   iFormService
	data       *payloads.FileWithComment
	modelError error
	inputs     []components.BlinkInput
	help       help.Model
	helpKeys   []key.Binding
}

// InitialForm инициализирует и возвращает форму с предопределенными полями ввода и привязками помощи по навигации с помощью клавиатуры.
func InitialForm(service iFormService, data *payloads.FileWithComment, lastModel tea.Model) Form {
	fp := components.NewTInput(pathPlaceholder, "", true)
	name := components.NewTInput(namePlaceholder, string(data.Name), false)
	comment := components.NewTArea(commentPlaceholder, data.Comment, false)

	m := Form{
		pService: service,
		data:     data,
		inputs: []components.BlinkInput{
			fp,
			name,
			comment,
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
		return m.updateFile()
	}
	// На комментарии нужно разрешать делать новую строку
	if s == "enter" && m.focusIndex == len(m.inputs)-1 {
		return m, models.UpdateInputs(msg, m.inputs)
	}
	m.focusIndex = models.IncrementCircleIndex(m.focusIndex, len(m.inputs), s)

	return m, models.GetCmds(m.inputs, m.focusIndex)
}

// updateFile обрабатывает ввод пользователя для создания или обновления записи файла, шифрует ее и возвращает соответствующую модель.
func (m Form) updateFile() (tea.Model, tea.Cmd) {
	m.data.Name = []byte(m.inputs[nameI].Value())
	m.data.Comment = m.inputs[commentI].Value()
	var err error
	m.data, err = m.pService.EncryptItem(m.data)
	if err == nil {
		if m.data.ID == "" {
			var encFilePath string
			encFilePath, err = m.pService.EncryptFile(m.inputs[filePathI].Value())
			if err == nil {
				defer os.Remove(encFilePath)
				err = m.pService.CreateFile(m.data, encFilePath)
			}
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

// View отображает форму на основе ее текущего состояния, включая входные данные, кнопки и ошибки, и возвращает визуализированную строку.
func (m Form) View() string {
	var b strings.Builder
	start := filePathI
	if m.data.ID != "" {
		fmt.Fprintf(&b, "%s\n\n", headerUpdateText)
		start = nameI
	} else {
		fmt.Fprintf(&b, "%s\n\n", headerNewText)
	}
	for i := start; i < len(m.inputs); i++ {
		b.WriteString(m.inputs[i].View())
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
