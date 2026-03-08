package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#FDFDFD")
	colorHighlight = lipgloss.Color("#EE6FF8")
	colorError     = lipgloss.Color("#EF4444")
	colorSuccess   = lipgloss.Color("#10B981")
	colorDim       = lipgloss.Color("#6B7280")
	colorText      = lipgloss.Color("#D1D5DB")

	// Layout Styles
	// We remove margins to allow truly filling the screen
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(colorSecondary).
			MarginBottom(1).
			Padding(0, 1).
			Bold(true)

	ListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			MarginRight(2)

	DetailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(1, 2)

	// Item Styles
	ItemStyle = lipgloss.NewStyle().
			Foreground(colorText).
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true).
				PaddingLeft(0).
				SetString("▶ ")

	DescStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	// Log Output Style
	OutputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			Foreground(colorText)

	// StatusBar Style
	StatusStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Background(colorPrimary).
			Padding(0, 1).
			MarginTop(1)
)
