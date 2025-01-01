package text

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
)

const (
	selectedTemplate = "%s\n\nТекст: %s\nКомментарий: %s\n\n%s"
	header           = "Список текстов"
	selectedHeader   = "Текст"
	newText          = "Новый текст"
	updateText       = "Обновить текст"
	deleteText       = "Удалить текст"
)

var (
	headerText         = style.HeaderStyle.Render(header)
	selectedHeaderText = style.HeaderStyle.Render(selectedHeader)
)

// iListService определяет интерфейс для управления текстовыми объектами, включая поиск, шифрование, дешифрование, создание, обновление и удаление.
type iListService interface {
	Get() ([]service.TextData, error)
	Delete(id string) error
}

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка паролей.
type List struct {
	list       list.Model
	pService   iListService
	modelError error
	selected   *service.TextData
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.TextService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(textService iListService) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: textService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", models.EscapeText), key.WithKeys("backspace", "esc")),
		},
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := make([]key.Binding, 3)
		binds[0] = key.NewBinding(key.WithHelp("n", newText), key.WithKeys("n"))
		binds[1] = key.NewBinding(key.WithHelp("u", updateText), key.WithKeys("u"))
		binds[2] = key.NewBinding(key.WithHelp("d", deleteText), key.WithKeys("d"))
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
			return m.newText()
		case "u":
			return m.updateText()
		case "d":
			return m.deleteText()
		}
	case tea.WindowSizeMsg:
		h, v := style.DocStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectItem Устанавливает текст в качестве текущего выбранного элемента в модели.
func (m List) selectItem() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.TextData)
	m.modelError = nil
	m.selected = &selected
	return m, nil
}

// newText переключает модель на форму создания нового текста и инициализирует форму значениями по умолчанию.
func (m List) newText() (tea.Model, tea.Cmd) {
	n := &payloads.TextWithComment{}
	newForm := InitialForm(service.NewDefaultTextService(), n)
	return newForm, newForm.Init()
}

// updateText переключает модель на форму обновления текста.
func (m List) updateText() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.TextData)
	selectedData := &selected.TextWithComment
	newForm := InitialForm(service.NewDefaultTextService(), selectedData)
	return newForm, newForm.Init()
}

// deleteText удаляет выбранный текст с помощью TextService и обновляет список модели.
// Возвращает обновленную модель и команду.
// Если во время удаления или обновления возникает ошибка, устанавливается modelError и не выдается команда.
func (m List) deleteText() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.TextData)
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
		}
	}

	return m, nil
}

// View генерирует и возвращает форматированное строковое представление списка, включая любой выбранный элемент или сообщение об ошибке.
func (m List) View() string {
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

// refresh обновить обновляет список, получая тексты из TextService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	texts, err := l.pService.Get()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(texts))
	for i, p := range texts {
		pl[i] = p
	}
	l.list.SetItems(pl)
	return nil
}

// renderSelected генерирует и возвращает форматированное строковое представление выбранного текста и его сведений.
func (l List) renderSelected() string {
	return fmt.Sprintf(selectedTemplate,
		selectedHeaderText,
		style.FocusedStyle.Render(string(l.selected.TextData)),
		l.selected.Comment,
		l.help.ShortHelpView(l.helpKeys))
}
