package card

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"passkeeper/internal/client/models"
	"passkeeper/internal/client/service"
	"passkeeper/internal/client/style"
	"passkeeper/internal/payloads"
)

const (
	selectedTemplate = "%s\n\nНомер: %s\nДата: %s\nДержатель: %s\nCVV: %s\nКомментарий: %s\n\n%s"
	header           = "Список карт"
	selectedHeader   = "Карта"
	newText          = "Новая карта"
	updateText       = "Обновить карту"
	deleteText       = "Удалить карту"
)

var (
	docStyle           = lipgloss.NewStyle()
	headerText         = style.HeaderStyle.Render(header)
	selectedHeaderText = style.HeaderStyle.Render(selectedHeader)
)

// iListService определяет интерфейс для обработки операций с данными карты, включая операции шифрования, дешифрования и CRUD.
type iListService interface {
	Get() ([]service.CardData, error)
	Delete(id string) error
}

// List представляет модель, управляющую отображением, состоянием и взаимодействием списка карт.
type List struct {
	list       list.Model
	pService   iListService
	modelError error
	selected   *service.CardData
	help       help.Model
	helpKeys   []key.Binding
}

// NewList инициализирует и возвращает новую модель списка, настроенную с использованием предоставленного service.CardService.
// Он устанавливает внутреннюю модель списка, клавиши справки и обновляет содержимое списка.
func NewList(cardService iListService) List {
	m := List{
		list:     list.New(nil, list.NewDefaultDelegate(), 0, 0),
		pService: cardService,
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
			return m.newCard()
		case "u":
			return m.updateCard()
		case "d":
			return m.deleteCard()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// selectItem Устанавливает карту в качестве текущего выбранного элемента в модели.
func (m List) selectItem() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.CardData)
	m.modelError = nil
	m.selected = &selected
	return m, nil
}

// newCard переключает модель на форму создания новой карты и инициализирует форму значениями по умолчанию.
func (m List) newCard() (tea.Model, tea.Cmd) {
	n := &payloads.CardWithComment{}
	newForm := InitialForm(service.NewDefaultCardService(), n)
	return newForm, newForm.Init()
}

// updateCard переключает модель на форму обновления карты.
func (m List) updateCard() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.CardData)
	selectedData := &selected.CardWithComment
	newForm := InitialForm(service.NewDefaultCardService(), selectedData)
	return newForm, newForm.Init()
}

// deleteCard удаляет выбранную карту с помощью CardService и обновляет список модели.
// Возвращает обновленную модель и команду.
// Если во время удаления или обновления возникает ошибка, устанавливается modelError и не выдается команда.
func (m List) deleteCard() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem().(service.CardData)
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

// refresh обновить обновляет список, получая карты из CardService и устанавливая их как элементы списка.
func (l *List) refresh() error {
	cards, err := l.pService.Get()
	if err != nil {
		return err
	}
	pl := make([]list.Item, len(cards))
	for i, p := range cards {
		pl[i] = p
	}
	l.list.SetItems(pl)
	return nil
}

// renderSelected генерирует и возвращает форматированное строковое представление выбранной карты и его сведений.
func (l List) renderSelected() string {
	return fmt.Sprintf(selectedTemplate,
		selectedHeaderText,
		style.FocusedStyle.Render(string(l.selected.Number)),
		style.FocusedStyle.Render(string(l.selected.Date)),
		style.FocusedStyle.Render(string(l.selected.Owner)),
		style.FocusedStyle.Render(string(l.selected.CVV)),
		l.selected.Comment,
		l.help.ShortHelpView(l.helpKeys))
}
