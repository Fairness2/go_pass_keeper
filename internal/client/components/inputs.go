package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/style"
)

type BlinkInput interface {
	Focus() tea.Cmd
	Blur()
	View() string
	Value() string
}

type TInput struct {
	textinput.Model
}

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

func NewTPass(placeholder, initialValue string, focus bool) *TInput {
	i := NewTInput(placeholder, initialValue, focus)
	i.EchoMode = textinput.EchoPassword
	i.EchoCharacter = '•'
	return i
}

func (i *TInput) Update(msg tea.Msg) (*TInput, tea.Cmd) {
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}

func (i *TInput) Blur() {
	i.Model.Blur()
	i.PromptStyle = style.NoStyle
	i.TextStyle = style.NoStyle
}

func (i *TInput) Focus() tea.Cmd {
	i.PromptStyle = style.FocusedStyle
	i.TextStyle = style.FocusedStyle
	return i.Model.Focus()
}

type TArea struct {
	textarea.Model
}

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

func (i *TArea) Update(msg tea.Msg) (*TArea, tea.Cmd) {
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}
