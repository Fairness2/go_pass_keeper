package style

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle        = FocusedStyle
	NoStyle            = lipgloss.NewStyle()
	ErrorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#fc0303"))
	ButtonFocusedStyle = FocusedStyle
	ButtonBlurredStyle = BlurredStyle
	HeaderStyle        = FocusedStyle

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
