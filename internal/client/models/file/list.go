package file

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"os"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
	"strings"
)

const (
	selectedTemplate = "%s\n\nНазвание: %s\nКомментарий: %s\n\n%s"
	header           = "Список файлов"
	selectedHeader   = "Файл"
	newText          = "Новый файлов"
	updateText       = "Обновить файлов"
	deleteText       = "Удалить файлов"
	pathText         = "Путь для загрузки"
	downloadText     = "Загрузить файл"
	changePathText   = "Изменить путь загрузки"
)

var (
	headerText         = style.HeaderStyle.Render(header)
	selectedHeaderText = style.HeaderStyle.Render(selectedHeader)
	filePathHeaderText = style.HeaderStyle.Render(pathText)
)

// processService определяет интерфейс для управления файлами, включая шифрование, дешифрование, создание, обновление и удаление.
// Он предоставляет методы для управления данными файла с соответствующими комментариями и поддерживает шифрование и дешифрование на уровне файла.
type iListService interface {
	Get() ([]service.FileData, error)
	Delete(id string) error
	DecryptFile(from io.Reader, dest io.Writer) error
	DownloadFile(id string, destFile string) error
}

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка файлов.
type List struct {
	list         list.Model
	pService     iListService
	modelError   error
	selected     *service.FileData
	help         help.Model
	helpKeys     []key.Binding
	helpPathKeys []key.Binding
	downloadPath string
	showPathForm bool
	pathInput    []components.BlinkInput
	focusIndex   int
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.FileService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(fileService iListService) List {
	initialDownloadPath := os.Getenv("HOME") + "/Downloads"
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: fileService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", models.BackText), key.WithKeys("backspace", "esc")),
			key.NewBinding(key.WithHelp("z", downloadText), key.WithKeys("z")),
			key.NewBinding(key.WithHelp("p", changePathText), key.WithKeys("p")),
		},
		helpPathKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", models.BackText), key.WithKeys("backspace", "esc")),
		},
		downloadPath: initialDownloadPath,
		pathInput:    []components.BlinkInput{components.NewTInput(pathText, initialDownloadPath, true)},
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := []key.Binding{
			key.NewBinding(key.WithHelp("n", newText), key.WithKeys("n")),
			key.NewBinding(key.WithHelp("u", updateText), key.WithKeys("u")),
			key.NewBinding(key.WithHelp("d", deleteText), key.WithKeys("d")),
			key.NewBinding(key.WithHelp("z", downloadText), key.WithKeys("z")),
			key.NewBinding(key.WithHelp("p", changePathText), key.WithKeys("p")),
		}
		return binds
	}

	if err := m.refresh(); err != nil {
		m.modelError = err
	}
	return m
}

// Init инициализирует модель List и возвращает команду для установки размера окна.
func (m List) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update обновляет модель на основе предоставленного сообщения и возвращает обновленную модель и необязательную команду.
func (m List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showPathForm {
		return m.pathUpdate(msg)
	}
	if m.selected != nil {
		return m.updateWhileSelected(msg)
	}
	return m.updateWhileNotSelected(msg)
}

// updateWhileNotSelected обрабатывает обновления для модели списка, когда ни один элемент не выбран, обрабатывая ввод пользователя и настройку размера окна.
func (m List) updateWhileNotSelected(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch expr := msg.String(); expr {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			return m.selectItem()
		case "n":
			return m.newFile()
		case "u":
			return m.updateFile()
		case "d":
			return m.deleteFile()
		case "z":
			return m.downloadFile()
		case "p":
			m.showPathForm = !m.showPathForm
			return m, nil
		}
	case tea.WindowSizeMsg:
		models.Resize(&m.list, msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectItem Устанавливает файл в качестве текущего выбранного элемента в модели.
func (m List) selectItem() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.FileData)
	m.modelError = nil
	m.selected = &selected
	return m, nil
}

// newFile переключает модель на форму создания нового файла и инициализирует форму значениями по умолчанию.
func (m List) newFile() (tea.Model, tea.Cmd) {
	n := &payloads.FileWithComment{}
	newForm := InitialForm(service.NewDefaultFileService(), n)
	return newForm, newForm.Init()
}

// updateFile переключает модель на форму обновления файла.
func (m List) updateFile() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.FileData)
	selectedData := &selected.FileWithComment
	newForm := InitialForm(service.NewDefaultFileService(), selectedData)
	return newForm, newForm.Init()
}

// deleteFile удаляет выбранную карту с помощью FileService и обновляет список модели.
// Возвращает обновленную модель и команду.
// Если во время удаления или обновления возникает ошибка, устанавливается modelError и не выдается команда.
func (m List) deleteFile() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.FileData)
	err := m.pService.Delete(selected.ID)
	if err != nil {
		m.modelError = err
		return m, nil
	}
	if err = m.refresh(); err != nil {
		m.modelError = err
	}
	return m, nil
}

// updateWhileSelected обрабатывает обновления модели при выбранном элементе.
func (m List) updateWhileSelected(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch expr := msg.String(); expr {
		case "esc", "backspace":
			m.selected = nil
			return m, nil
		case "z":
			return m.downloadFile()
		case "p":
			m.showPathForm = !m.showPathForm
			return m, nil
		}
	}

	return m, nil
}

// View генерирует и возвращает форматированное строковое представление списка, включая любой выбранный элемент или сообщение об ошибке.
func (m List) View() string {
	if m.showPathForm {
		return m.renderPath()
	}
	if m.selected != nil {
		return m.renderSelected()
	}
	return m.renderList()
}

// renderList генерирует форматированное строковое представление списка, включая сообщение об ошибке, если оно существует.
func (m List) renderList() string {
	errorStr := ""
	if m.modelError != nil {
		errorStr = style.ErrorStyle.Render(m.modelError.Error())
	}
	return fmt.Sprintf(models.ListTemplate, headerText, errorStr, m.list.View())
}

// refresh обновить обновляет список, получая файлы из FileService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	files, err := l.pService.Get()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(files))
	for i, p := range files {
		pl[i] = p
	}
	l.list.SetItems(pl)
	return nil
}

// renderSelected генерирует и возвращает форматированное строковое представление выбранного файла и его сведений.
func (l List) renderSelected() string {
	return fmt.Sprintf(selectedTemplate,
		selectedHeaderText,
		style.FocusedStyle.Render(string(l.selected.Name)),
		l.selected.Comment,
		l.help.ShortHelpView(l.helpKeys))
}

// downloadFile загружает выбранный файл, расшифровывает его и сохраняет по указанному пути загрузки, обрабатывая ошибки, если таковые имеются.
func (l List) downloadFile() (tea.Model, tea.Cmd) {
	selected := l.list.SelectedItem().(service.FileData)
	temp, err := os.CreateTemp(os.TempDir(), "file-*.zip")
	if err != nil {
		l.modelError = err
		return l, nil
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()
	destFile, err := os.Create(fmt.Sprintf("%s/%s", l.downloadPath, string(selected.Name))) // TODO
	if err != nil {
		l.modelError = err
		return l, nil
	}
	defer destFile.Close()

	if err = l.pService.DownloadFile(selected.ID, temp.Name()); err != nil {
		l.modelError = err
		return l, nil
	}
	if err = l.pService.DecryptFile(temp, destFile); err != nil {
		l.modelError = err
		return l, nil
	}

	return l, nil
}

// renderSelected генерирует и возвращает форматированное строковое представление заполнения пути загрузки.
func (l List) renderPath() string {
	var b strings.Builder
	b.WriteString(filePathHeaderText)
	b.WriteString("\n")
	for _, i := range l.pathInput {
		b.WriteString(i.View())
		b.WriteString("\n")
	}
	button := &blurredButton
	if l.focusIndex == len(l.pathInput) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "%s\n\n", *button)
	b.WriteString(l.help.ShortHelpView(l.helpPathKeys))
	return b.String()
}

// pathUpdate обрабатывает сообщения для обновления состояния списка при заполнении пути загрузки, управляет фокусом и обрабатывает ввод для формы.
func (l List) pathUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return l, tea.Quit
		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			if s == "enter" && l.focusIndex == len(l.pathInput) {
				l.downloadPath = l.pathInput[0].Value()
				l.showPathForm = false
				return l, nil
			}
			// На комментарии нужно разрешать делать новую строку
			if s == "enter" && l.focusIndex == len(l.pathInput)-1 {
				break
			}
			l.focusIndex = models.IncrementCircleIndex(l.focusIndex, len(l.pathInput), s)

			return l, models.GetCmds(l.pathInput, l.focusIndex)
		}
	}
	// Handle character input and blinking
	cmd := models.UpdateInputs(msg, l.pathInput)
	return l, cmd
}
