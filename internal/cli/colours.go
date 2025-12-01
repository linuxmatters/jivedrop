package cli

import "github.com/charmbracelet/lipgloss"

// Disco ball colour palette 🪩
// Cool blues, cyans, purples and silvers - like light reflecting off a glitter ball
var (
	PrimaryColor   = lipgloss.Color("#00BFFF") // Deep sky blue - core disco reflection
	AccentColor    = lipgloss.Color("#00FFFF") // Electric cyan - sparkling highlights
	SuccessColor   = lipgloss.Color("#00CED1") // Dark turquoise - cool success
	MutedColor     = lipgloss.Color("#778899") // Light slate gray
	HighlightColor = lipgloss.Color("#E0E0E0") // Silver/white - mirror reflection
	TextColor      = lipgloss.Color("#FFFFFF") // White
	ErrorColor     = lipgloss.Color("#DA70D6") // Orchid - distinct but cool
	SecondaryColor = lipgloss.Color("#9370DB") // Medium purple - disco purple tones
	BorderColor    = lipgloss.Color("#00BFFF") // Deep sky blue - glittery border

	// Disco ball gradient colours (indigo → white)
	GradientIndigo = lipgloss.Color("#4B0082") // Deep indigo
	GradientWhite  = lipgloss.Color("#E0E0E0") // Silver/white
)
