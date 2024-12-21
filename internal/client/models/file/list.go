package file

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
	"strings"
)

var (
	docStyle           = lipgloss.NewStyle()
	headerText         = style.HeaderStyle.Render("Список файлов")
	selectedHeaderText = style.HeaderStyle.Render("Файл")
)

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка файлов.
type List struct {
	list       list.Model
	pService   *service.FileService
	modelError error
	selected   *service.FileData
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.CardService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(fileService *service.FileService) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: fileService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", "Выход"), key.WithKeys("backspace", "esc")),
			key.NewBinding(key.WithHelp("z", "Загрузить файл"), key.WithKeys("z")),
		},
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := []key.Binding{
			key.NewBinding(key.WithHelp("n", "Новый файл"), key.WithKeys("n")),
			key.NewBinding(key.WithHelp("u", "Обновить файл"), key.WithKeys("u")),
			key.NewBinding(key.WithHelp("d", "Удалить файл"), key.WithKeys("d")),
			key.NewBinding(key.WithHelp("z", "Загрузить файл"), key.WithKeys("z")),
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
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
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
	newForm := InitialForm(m.pService, n)
	return newForm, newForm.Init()
}

// updateFile переключает модель на форму обновления файла.
func (m List) updateFile() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.FileData)
	selectedData := &selected.FileWithComment
	newForm := InitialForm(m.pService, selectedData)
	return newForm, newForm.Init()
}

// deleteFile удаляет выбранную карту с помощью FileService и обновляет список модели.
// Возвращает обновленную модель и команду.
// Если во время удаления или обновления возникает ошибка, устанавливается modelError и не выдается команда.
func (m List) deleteFile() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.FileData)
	err := m.pService.DeleteFile(selected.ID)
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
		}
	}

	return m, nil
}

// View генерирует и возвращает форматированное строковое представление списка, включая любой выбранный элемент или сообщение об ошибке.
func (m List) View() string {
	if m.selected != nil {
		return m.renderSelected()
	}
	var b strings.Builder
	b.WriteString(headerText)
	//b.WriteString("\n")
	if m.modelError != nil {
		fmt.Fprintf(&b, "%s", style.ErrorStyle.Render(m.modelError.Error()))
	}
	b.WriteString(m.list.View())
	return b.String()
}

// refresh обновить обновляет список, получая файлы из FileService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	files, err := l.pService.GetFiles()
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
	var b strings.Builder
	b.WriteString(selectedHeaderText)
	b.WriteString("\n")
	fmt.Fprintf(&b,
		"\nНазвание: %s\nКомментарий: %s\n",
		style.FocusedStyle.Render(string(l.selected.Name)),
		l.selected.Comment)
	b.WriteString("\n")
	b.WriteString(l.help.ShortHelpView(l.helpKeys))
	return b.String()
}

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
	destFile, err := os.Create(fmt.Sprintf("%s/%s", "/Users/konstantinkuzminyh/sites/go_pass_keeper/download", string(selected.Name)))
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
