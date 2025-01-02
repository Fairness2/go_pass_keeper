package password

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
	selectedTemplate = "%s\n\nЛогин: %s\nПароль: %s\nДомен: %s\nКомментарий: %s\n\n%s"
	header           = "Список паролей"
	selectedHeader   = "Пароль"
	newText          = "Новый пароль"
	updateText       = "Обновить пароль"
	deleteText       = "Удалить пароль"
)

var (
	headerText         = style.HeaderStyle.Render(header)
	selectedHeaderText = style.HeaderStyle.Render(selectedHeader)
)

// iListService определяет интерфейс для управления данными пароля, включая операции извлечения, шифрования, расшифровки, создания, обновления и удаления.
type iListService interface {
	Get() ([]service.PassData, error)
	Delete(id string) error
}

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка паролей.
type List struct {
	models.Backable
	list       *list.Model
	pService   iListService
	modelError error
	selected   *service.PassData
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.PasswordService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(passwordService iListService, lastModel tea.Model) List {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	m := List{
		list:     &l,
		pService: passwordService,
		help:     help.New(),
		helpKeys: []key.Binding{
			key.NewBinding(key.WithHelp("esc, backspace", models.EscapeText), key.WithKeys("backspace", "esc")),
		},
		Backable: models.NewBackable(lastModel),
	}
	m.list.SetShowTitle(false)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		binds := make([]key.Binding, 4)
		binds[0] = key.NewBinding(key.WithHelp("n", newText), key.WithKeys("n"))
		binds[1] = key.NewBinding(key.WithHelp("u", updateText), key.WithKeys("u"))
		binds[2] = key.NewBinding(key.WithHelp("d", deleteText), key.WithKeys("d"))
		binds[3] = key.NewBinding(key.WithHelp("esc", models.BackText), key.WithKeys("esc"))
		return binds
	}
	return m
}

// Init инициализирует модель List и возвращает команду для установки размера окна.
func (m List) Init() tea.Cmd {
	if err := m.refresh(); err != nil {
		m.modelError = err
	}
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
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.Back()
		case "enter":
			return m.selectItem()
		case "n":
			return m.newPassword()
		case "u":
			return m.updatePassword()
		case "d":
			return m.deletePassword()
		}
	case tea.WindowSizeMsg:
		models.Resize(m.list, msg)
	}

	var cmd tea.Cmd
	*m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectItem расшифровывает пароль выбранного элемента, обновляет его состояние и устанавливает его в качестве текущего выбранного элемента в модели.
func (m List) selectItem() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.PassData)
	m.modelError = nil
	m.selected = &selected
	return m, nil
}

// newPassword переключает модель на форму создания нового пароля и инициализирует форму значениями по умолчанию.
func (m List) newPassword() (tea.Model, tea.Cmd) {
	n := &payloads.PasswordWithComment{}
	newForm := InitialForm(service.NewDefaultPasswordService(), n, m)
	return newForm, newForm.Init()
}

// updatePassword переключает модель на форму обновления пароля, при необходимости расшифровывая выбранный пароль.
func (m List) updatePassword() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.PassData)
	selectedData := &selected.PasswordWithComment
	newForm := InitialForm(service.NewDefaultPasswordService(), selectedData, m)
	return newForm, newForm.Init()
}

// deletePassword удаляет выбранный пароль с помощью PasswordService и обновляет список модели.
// Возвращает обновленную модель и команду.
// Если во время удаления или обновления возникает ошибка, устанавливается modelError и не выдается команда.
func (m List) deletePassword() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.PassData)
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

// refresh обновить обновляет список, получая пароли из PasswordService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	passwords, err := l.pService.Get()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(passwords))
	for i, p := range passwords {
		pl[i] = p
	}
	l.list.SetItems(pl)
	return nil
}

// renderSelected генерирует и возвращает форматированное строковое представление выбранного пароля и его сведений.
func (l List) renderSelected() string {
	return fmt.Sprintf(selectedTemplate,
		selectedHeaderText,
		style.FocusedStyle.Render(string(l.selected.Username)),
		style.FocusedStyle.Render(string(l.selected.Password.Password)),
		style.FocusedStyle.Render(l.selected.Domen),
		l.selected.Comment,
		l.help.ShortHelpView(l.helpKeys))
}
