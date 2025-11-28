package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Disco ball colour palette 🪩
// Cool blues, cyans, purples and silvers - like light reflecting off a glitter ball
var (
	primaryColor   = lipgloss.Color("#00BFFF") // Deep sky blue - core disco reflection
	accentColor    = lipgloss.Color("#00FFFF") // Electric cyan - sparkling highlights
	successColor   = lipgloss.Color("#00CED1") // Dark turquoise - cool success
	mutedColor     = lipgloss.Color("#778899") // Light slate gray
	highlightColor = lipgloss.Color("#E0E0E0") // Silver/white - mirror reflection
	textColor      = lipgloss.Color("#FFFFFF") // White
	errorColor     = lipgloss.Color("#DA70D6") // Orchid - distinct but cool
	secondaryColor = lipgloss.Color("#9370DB") // Medium purple - disco purple tones
	borderColor    = lipgloss.Color("#00BFFF") // Deep sky blue - glittery border

	// Disco ball gradient colours (indigo → purple → cyan → white)
	gradientIndigo = lipgloss.Color("#4B0082") // Deep indigo
	gradientPurple = lipgloss.Color("#9370DB") // Medium purple
	gradientCyan   = lipgloss.Color("#00FFFF") // Electric cyan
	gradientWhite  = lipgloss.Color("#E0E0E0") // Silver/white
)

// Header style for section titles
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(primaryColor).
	MarginBottom(1)

// Accent style for highlighted values
var accentStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(accentColor)

// Success message style
var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(successColor)

// Error message style
var errorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(errorColor)

// Highlight style for important values
var highlightStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(highlightColor)

// Key-value pair styles
var keyStyle = lipgloss.NewStyle().
	Foreground(mutedColor)

var valueStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(textColor)

// Muted text style
var mutedStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Italic(true)

// Progress bar style
var progressBarStyle = lipgloss.NewStyle().
	Foreground(accentColor).
	Bold(true)

// Box style for framed content
var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(borderColor).
	Padding(1, 2).
	MarginTop(1).
	MarginBottom(1)
