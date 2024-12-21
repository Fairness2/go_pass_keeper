package style

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))     // Определяет стиль для сфокусированных элементов пользовательского интерфейса с цветом переднего плана «205».
	BlurredStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))     // Определяет стиль с помощью тусклого цвета переднего плана, используемого для несфокусированных элементов.
	CursorStyle        = FocusedStyle                                              // Определяет стиль курсора, наследуемый от конфигурации FocusedStyle.
	NoStyle            = lipgloss.NewStyle()                                       // Определяет стиль для обычных элементов.
	ErrorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#fc0303")) // Определяет стиль отображения сообщений об ошибках с красным цветом переднего плана.
	ButtonFocusedStyle = FocusedStyle                                              // Определяет стиль фокуса, используемый для рендеринга кнопок с фокусом.
	ButtonBlurredStyle = BlurredStyle                                              // Определяет стиль неактивной кнопки.
	HeaderStyle        = FocusedStyle                                              // Определяет стиль заголовков, используя FocusedStyle по умолчанию.

	// TextAreaFocused определяет стиль, применяемый к текстовой области, когда она находится в фокусе, включая текст, курсор и визуальные элементы.
	TextAreaFocused = textarea.Style{
		Base:             NoStyle,
		CursorLine:       NoStyle.Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
		CursorLineNumber: NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "240"}),
		EndOfBuffer:      NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      NoStyle.Foreground(lipgloss.Color("240")),
		Prompt:           NoStyle.Foreground(lipgloss.Color("7")),
		Text:             FocusedStyle,
	}

	// TextAreaBlurred определяет стиль, применяемый к текстовой области, когда она не в фокусе, включая текст, курсор и визуальные элементы.
	TextAreaBlurred = textarea.Style{
		Base:             NoStyle,
		CursorLine:       NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
		CursorLineNumber: NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		EndOfBuffer:      NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       NoStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      NoStyle.Foreground(lipgloss.Color("240")),
		Prompt:           NoStyle.Foreground(lipgloss.Color("7")),
		Text:             BlurredStyle,
	}
)
