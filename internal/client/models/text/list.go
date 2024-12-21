package text

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
	"strings"
)

var (
	docStyle           = lipgloss.NewStyle()
	headerText         = style.HeaderStyle.Render("Список текстов")
	selectedHeaderText = style.HeaderStyle.Render("Текст")
)

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка паролей.
type List struct {
	list       list.Model
	pService   *service.CRUDService[*payloads.TextWithComment, service.TextData]
	modelError error
	selected   *service.TextData
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.TextService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(textService *service.CRUDService[*payloads.TextWithComment, service.TextData]) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: textService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", "Выход"), key.WithKeys("backspace", "esc")),
		},
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := make([]key.Binding, 3)
		binds[0] = key.NewBinding(key.WithHelp("n", "Новый текст"), key.WithKeys("n"))
		binds[1] = key.NewBinding(key.WithHelp("u", "Обновить текст"), key.WithKeys("u"))
		binds[2] = key.NewBinding(key.WithHelp("d", "Удалить текст"), key.WithKeys("d"))
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
		h, v := docStyle.GetFrameSize()
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
	newForm := InitialForm(m.pService, n)
	return newForm, newForm.Init()
}

// updateText переключает модель на форму обновления текста.
func (m List) updateText() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.TextData)
	selectedData := &selected.TextWithComment
	newForm := InitialForm(m.pService, selectedData)
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
	var b strings.Builder
	b.WriteString(headerText)
	//b.WriteString("\n")
	if m.modelError != nil {
		fmt.Fprintf(&b, "\n%s\n", style.ErrorStyle.Render(m.modelError.Error()))
	}
	b.WriteString(m.list.View())
	return b.String()
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
	var b strings.Builder
	b.WriteString(selectedHeaderText)
	b.WriteString("\n")
	fmt.Fprintf(&b,
		"\nТекст: %s\nКомментарий: %s\n",
		style.FocusedStyle.Render(string(l.selected.TextData)),
		l.selected.Comment)
	b.WriteString("\n")
	b.WriteString(l.help.ShortHelpView(l.helpKeys))
	return b.String()
}
