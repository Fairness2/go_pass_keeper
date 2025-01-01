package models

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/components"
	"passkeeper/internal/client/style"
	"strings"
)

const (
	EscapeText     = "Выход"
	NavigationText = "Переход по форме"
	OKText         = "Принять"
	SaveText       = "[ Сохранить ]"
	BackText       = "Назад"

	ListTemplate = "%s\n%s\n%s"
	FormTemplate = "%s\n\n%s\n%s%s\n\n%s"
)

var (
	BaseFormHelp = []key.Binding{
		key.NewBinding(key.WithHelp("ctrl+c, esc", EscapeText), key.WithKeys("ctrl+c", "esc")),
		key.NewBinding(key.WithHelp("tab, shift+tab, up, down", NavigationText), key.WithKeys("tab", "shift+tab", "up", "down")),
		key.NewBinding(key.WithHelp("enter", OKText), key.WithKeys("enter")),
	}
)

// IncrementCircleIndex циклически корректирует значение индекса на основе заданного ключа и общего количества полей.
// Если клавиша «вверх» или «shift+tab», индекс уменьшается, в противном случае — увеличивается.
// Результат циклически изменяется в диапазоне от 0 до fieldLen включительно.
func IncrementCircleIndex(index, fieldsLen int, key string) int {
	// Cycle indexes
	if key == "up" || key == "shift+tab" {
		index--
	} else {
		index++
	}
	if index > fieldsLen {
		index = 0
	} else if index < 0 {
		index = fieldsLen
	}
	return index
}

// GetCmds генерирует пакетную команду для обновления состояния фокуса входных данных формы на основе текущего индекса фокуса.
func GetCmds(inputs []components.BlinkInput, focusIndex int) tea.Cmd {
	cmds := make([]tea.Cmd, len(inputs))
	for i, input := range inputs {
		if focusIndex == i {
			cmds[i] = input.Focus()
		} else {
			input.Blur()
		}
	}
	return tea.Batch(cmds...)
}

type FormViewConfig struct {
	HeaderNewText    string
	HeaderUpdateText string
	BlurredButton    string
	FocusedButton    string
}

func FormView(cnf *FormViewConfig, dataID string, inputs []components.BlinkInput, focusIndex int, modelError error, help string) string {
	// Выбираем текст заголовка
	h := cnf.HeaderNewText
	if dataID != "" {
		h = cnf.HeaderUpdateText
	}

	// Составляем текст импутов
	var b strings.Builder
	for _, input := range inputs {
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	// Составляем текст ошибки
	var errStr string
	if modelError != nil {
		errStr = style.ErrorStyle.Render(modelError.Error())
	}

	// Составляем текст кнопки
	button := cnf.BlurredButton
	if focusIndex == len(inputs) {
		button = cnf.FocusedButton
	}

	return fmt.Sprintf(FormTemplate, h, b.String(), errStr, button, help)
}

// UpdateInputs обновляет все входные компоненты в форме на основе предоставленного сообщения и возвращает пакетную команду для обновлений.
func UpdateInputs(msg tea.Msg, inputs []components.BlinkInput) tea.Cmd {
	cmds := make([]tea.Cmd, len(inputs))
	for i, input := range inputs {
		switch r := input.(type) {
		case *components.TInput:
			inputs[i], cmds[i] = r.Update(msg)
		case *components.TArea:
			inputs[i], cmds[i] = r.Update(msg)
		}
	}
	return tea.Batch(cmds...)
}
