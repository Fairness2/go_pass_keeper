package password

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
	headerText         = style.HeaderStyle.Render("Список паролей")
	selectedHeaderText = style.HeaderStyle.Render("Пароль")
)

// PassData представляет собой структуру данных пароля с дополнительным комментарием и расшифрованным состоянием.
type PassData struct {
	payloads.PasswordWithComment
	isDecrypted bool
}

func (i PassData) Title() string       { return i.Domen }
func (i PassData) Description() string { return i.Comment }
func (i PassData) FilterValue() string { return i.Domen }

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка паролей.
type List struct {
	list       list.Model
	pService   *service.PasswordService
	modelError error
	selected   *payloads.PasswordWithComment
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.PasswordService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(passwordService *service.PasswordService) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: passwordService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", "Выход"), key.WithKeys("backspace", "esc")),
		},
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := make([]key.Binding, 2)
		binds[0] = key.NewBinding(key.WithHelp("n", "Новый пароль"), key.WithKeys("n"))
		binds[1] = key.NewBinding(key.WithHelp("u", "Обновить пароль"), key.WithKeys("u"))
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
			return m.newPassword()
		case "u":
			return m.updatePassword()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectItem расшифровывает пароль выбранного элемента, обновляет его состояние и устанавливает его в качестве текущего выбранного элемента в модели.
func (m List) selectItem() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(PassData)
	pas, err := m.pService.DecryptPassword(&selected.PasswordWithComment)
	if err != nil {
		m.modelError = err
		return m, nil
	}
	selected.isDecrypted = true
	m.modelError = nil
	m.selected = pas
	return m, nil
}

// newPassword переключает модель на форму создания нового пароля и инициализирует форму значениями по умолчанию.
func (m List) newPassword() (tea.Model, tea.Cmd) {
	n := &payloads.PasswordWithComment{}
	newForm := InitialForm(m.pService, n)
	return newForm, newForm.Init()
}

// updatePassword переключает модель на форму обновления пароля, при необходимости расшифровывая выбранный пароль.
func (m List) updatePassword() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(PassData)
	var selectedData *payloads.PasswordWithComment
	var err error
	if selected.isDecrypted {
		selectedData = &selected.PasswordWithComment
	} else {
		selectedData, err = m.pService.DecryptPassword(&selected.PasswordWithComment)
		if err != nil {
			m.modelError = err
			return m, nil
		}
	}
	newForm := InitialForm(m.pService, selectedData)
	return newForm, newForm.Init()
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

// refresh обновить обновляет список, получая пароли из PasswordService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	passwords, err := l.pService.GetPasswords()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(passwords))
	for i, p := range passwords {
		pl[i] = PassData{
			PasswordWithComment: p,
		}
	}
	l.list.SetItems(pl)
	return nil
}

// renderSelected генерирует и возвращает форматированное строковое представление выбранного пароля и его сведений.
func (l List) renderSelected() string {
	var b strings.Builder
	b.WriteString(selectedHeaderText)
	b.WriteString("\n")
	fmt.Fprintf(&b,
		"\nЛогин: %s\nПароль: %s\nДомен: %s\nКомментарий: %s\n",
		style.FocusedStyle.Render(string(l.selected.Username)),
		style.FocusedStyle.Render(string(l.selected.Password.Password)),
		style.FocusedStyle.Render(l.selected.Domen),
		l.selected.Comment)
	b.WriteString("\n")
	b.WriteString(l.help.ShortHelpView(l.helpKeys))
	return b.String()
}
