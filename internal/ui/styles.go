package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/linuxmatters/jivedrop/internal/cli"
)

// Import shared colour palette from cli package
var (
	primaryColor = cli.PrimaryColor
	accentColor  = cli.AccentColor
	mutedColor   = cli.MutedColor

	// Disco ball gradient colours
	gradientIndigo = cli.GradientIndigo
	gradientWhite  = cli.GradientWhite
)

// Header style for section titles
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(primaryColor).
	MarginBottom(1)

// Spinner style for the encoding progress indicator
var spinnerStyle = lipgloss.NewStyle().
	Foreground(accentColor)

// Shared styles from cli package
var (
	successStyle   = cli.SuccessStyle
	errorStyle     = cli.ErrorStyle
	highlightStyle = cli.HighlightStyle
	keyStyle       = cli.KeyStyle
	valueStyle     = cli.ValueStyle
	boxStyle       = cli.BoxStyle
)

// Muted text style (no cli equivalent)
var mutedStyle = lipgloss.NewStyle().
	Foreground(mutedColor).
	Italic(true)

// clockStyle renders the elapsed/remaining MM:SS values in white, matching the
// unbolded weight of the media-player stats row.
var clockStyle = lipgloss.NewStyle().
	Foreground(cli.TextColor)

// frameWidth fixes the lipgloss style width (content area plus the box's 2-cell
// horizontal padding each side) shared by every framed view, so the progress,
// complete and error boxes match in outer width. Derived from the widest
// progress row (the 40-cell bar + "  " + "NNN%" = 46) plus 4 cells of padding.
const frameWidth = 50

// frameStyle is the shared box used by every framed view at a fixed width.
var frameStyle = boxStyle.Width(frameWidth)

// boltStyle renders the speed lightning bolt in a warm gold against the cool
// palette, giving it an understated "energy" cue.
var boltStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFD166"))
