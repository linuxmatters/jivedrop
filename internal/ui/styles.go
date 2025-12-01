package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/linuxmatters/jivedrop/internal/cli"
)

// Import shared colour palette from cli package
var (
	primaryColor   = cli.PrimaryColor
	accentColor    = cli.AccentColor
	successColor   = cli.SuccessColor
	mutedColor     = cli.MutedColor
	highlightColor = cli.HighlightColor
	textColor      = cli.TextColor
	errorColor     = cli.ErrorColor
	secondaryColor = cli.SecondaryColor
	borderColor    = cli.BorderColor

	// Disco ball gradient colours
	gradientIndigo = cli.GradientIndigo
	gradientPurple = cli.GradientPurple
	gradientCyan   = cli.GradientCyan
	gradientWhite  = cli.GradientWhite
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
