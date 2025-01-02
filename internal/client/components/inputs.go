package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/style"
)

// BlinkInput — это интерфейс, представляющий фокусируемый элемент ввода с методами фокусировки, размытия и получения его значения.
type BlinkInput interface {
	Focus() tea.Cmd
	Blur()
	View() string
	Value() string
}

// TInput обертывает textinput.Model, чтобы расширить его функциональность для сфокусированного и размытого стиля ввода.
type TInput struct {
	textinput.Model
}

// NewTInput инициализирует и возвращает указатель на TInput, настроенный с помощью заполнителя, начального значения и состояния фокуса.
func NewTInput(placeholder, initialValue string, focus bool) *TInput {
	i := textinput.New()
	i.Cursor.Style = style.CursorStyle
	i.Placeholder = placeholder
	i.SetValue(initialValue)
	if focus {
		i.Focus()
		i.PromptStyle = style.FocusedStyle
		i.TextStyle = style.FocusedStyle
	} else {
		i.Blur()
		i.PromptStyle = style.NoStyle
		i.TextStyle = style.NoStyle
	}
	return &TInput{i}
}

// NewTPass создает новый экземпляр TInput, настроенный для ввода пароля, с указанным заполнителем, начальным значением и состоянием фокуса.
// Поле ввода использует скрытый режим эха и символ эха '•'.
func NewTPass(placeholder, initialValue string, focus bool) *TInput {
	i := NewTInput(placeholder, initialValue, focus)
	i.EchoMode = textinput.EchoPassword
	i.EchoCharacter = '•'
	return i
}

// Update обновляет состояние TInput на основе входящего сообщения и возвращает обновленный TInput и tea.Cmd.
func (i *TInput) Update(msg tea.Msg) (*TInput, tea.Cmd) {
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}

// Blur удаляет сфокусированное состояние ввода, размывая его и сбрасывая стили подсказки и текста без стилей.
func (i *TInput) Blur() {
	i.Model.Blur()
	i.PromptStyle = style.NoStyle
	i.TextStyle = style.NoStyle
}

// Focus устанавливает ввод в сфокусированное состояние, применяет сфокусированные стили к подсказке и тексту и возвращает tea.Cmd.
func (i *TInput) Focus() tea.Cmd {
	i.PromptStyle = style.FocusedStyle
	i.TextStyle = style.FocusedStyle
	return i.Model.Focus()
}

// TArea — это оболочка textarea.Model, предоставляющая дополнительные функциональные возможности для компонентов текстовой области в пользовательском интерфейсе.
type TArea struct {
	textarea.Model
}

// NewTArea создает новый экземпляр TArea с указанным заполнителем, начальным значением и состоянием фокуса.
func NewTArea(placeholder, initialValue string, focus bool) *TArea {
	a := textarea.New()
	a.Cursor.Style = style.CursorStyle
	a.Placeholder = placeholder
	a.SetValue(initialValue)
	if focus {
		a.Focus()
		a.FocusedStyle = style.TextAreaFocused
	} else {
		a.Blur()
		a.BlurredStyle = style.TextAreaBlurred
	}
	return &TArea{a}
}

// Update обрабатывает данное сообщение для обновления состояния TArea и возвращает обновленный экземпляр TArea и команду.
func (i *TArea) Update(msg tea.Msg) (*TArea, tea.Cmd) {
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}
