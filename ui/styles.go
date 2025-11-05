package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#FF79C6")
	successColor   = lipgloss.Color("#50FA7B")
	warningColor   = lipgloss.Color("#FFB86C")
	errorColor     = lipgloss.Color("#FF5555")
	infoColor      = lipgloss.Color("#8BE9FD")
	mutedColor     = lipgloss.Color("#6272A4")
	backgroundColor = lipgloss.Color("#282A36")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true).
			BorderForeground(primaryColor).
			Padding(0, 1)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(primaryColor).
			Padding(0, 2).
			MarginRight(1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2).
				MarginRight(1)

	// Card style
	CardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginBottom(1)

	// Status styles
	RunningStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	StoppedStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	PausedStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Info styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			MaxWidth(100).
			Width(100)

	// Footer style
	FooterStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(mutedColor).
			Padding(1, 2)
)
