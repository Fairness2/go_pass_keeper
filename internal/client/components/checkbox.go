package components

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"passkeeper/internal/client/style"
	"strings"
)

// Checkbox представляет собой выбираемый список текстовых опций с одним активным выбором.
type Checkbox struct {
	choice int
	inputs []string
}

// NewCheckbox создает и возвращает новый экземпляр Checkbox с указанным начальным индексом выбора и параметрами ввода.
func NewCheckbox(initIndex int, inputs ...string) *Checkbox {
	return &Checkbox{
		choice: initIndex,
		inputs: inputs,
	}
}

// GetChoice возвращает индекс текущего активного выбора из параметров флажка.
func (c *Checkbox) GetChoice() int {
	return c.choice
}

// Update обрабатывает ввод пользователя, обновляет состояние флажка и возвращает потенциальную команду для дальнейшей обработки.
func (c *Checkbox) Update(msg tea.Msg) (*Checkbox, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return c, tea.Quit
		case "j", "down":
			c.choice++
			if c.choice > len(c.inputs)-1 {
				c.choice = 0
			}
		case "k", "up":
			c.choice--
			if c.choice < 0 {
				c.choice = len(c.inputs) - 1
			}
		}
	}

	return c, nil
}

// checkbox отображает элемент флажка как строку, используя другой стиль в зависимости от того, отмечен он или нет.
func (c Checkbox) checkbox(label string, checked bool) string {
	if checked {
		return style.FocusedStyle.Render(fmt.Sprintf("[x] %s", label))
	}
	return style.BlurredStyle.Render(fmt.Sprintf("[ ] %s", label))
}

// View генерирует строковое представление компонента флажка, отображая все параметры и указывая выбранный.
func (c Checkbox) View() string {
	var b strings.Builder
	for i, v := range c.inputs {
		b.WriteString(c.checkbox(v, i == c.choice))
		if i < len(c.inputs)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
