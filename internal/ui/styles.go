package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - matching Jivefire/Jivedrop aesthetic
var (
	primaryColor   = lipgloss.Color("#A40000") // Jivedrop red
	accentColor    = lipgloss.Color("#FFA500") // Orange/gold
	successColor   = lipgloss.Color("#00AA00") // Green
	mutedColor     = lipgloss.Color("#888888") // Gray
	highlightColor = lipgloss.Color("#FFFF00") // Yellow
	textColor      = lipgloss.Color("#FFFFFF") // White
	progressColor  = lipgloss.Color("#FFA500") // Orange for progress bar
)

// Header style for section titles
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(accentColor).
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
	Foreground(primaryColor)

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
	Foreground(progressColor).
	Bold(true)

// Box style for framed content
var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor).
	Padding(1, 2).
	MarginTop(1).
	MarginBottom(1)
