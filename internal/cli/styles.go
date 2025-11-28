package cli

import (
	"fmt"
	"os"

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
)

// Styles
var (
	// Title style - bold blue with disco ball emoji
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle style - muted slate gray
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Section header style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor).
			MarginTop(1).
			MarginBottom(1)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)

	// Highlight style for important values
	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor)

	// Key-value pair styles
	KeyStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	ValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor)

	// Box style for framed content
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

// PrintBanner prints the application banner
func PrintBanner() {
	banner := TitleStyle.Render("Jivedrop 🪩")
	subtitle := SubtitleStyle.Render("Drop your podcast .wav into a shiny MP3 with metadata, cover art, and all.")
	fmt.Println(banner)
	fmt.Println(subtitle)
	fmt.Println()
}

// PrintVersion prints version information
func PrintVersion(version string) {
	fmt.Println(TitleStyle.Render("Jivedrop 🪩"))
	fmt.Printf("%s %s\n", KeyStyle.Render("Version:"), ValueStyle.Render(version))
	fmt.Println()
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ErrorStyle.Render("Error:"), message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render("Warning:"), message)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", SuccessStyle.Render("✓"), message)
}

// PrintInfo prints an informational message
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", KeyStyle.Render("•"), message)
}

// PrintSection prints a section header
func PrintSection(title string) {
	fmt.Println()
	fmt.Println(HeaderStyle.Render(title))
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
	fmt.Printf("%s %s\n", KeyStyle.Render(key+":"), ValueStyle.Render(value))
}

// PrintLabelValue prints a label with muted style and a value
// Used for summary output like "Episode: 67 - Title"
func PrintLabelValue(label, value string) {
	fmt.Printf("%s %s\n", KeyStyle.Render(label), value)
}

// PrintSuccessLabel prints a success checkmark with a muted label and value
func PrintSuccessLabel(label, value string) {
	fmt.Printf("%s %s %s\n", SuccessStyle.Render("\u2713"), KeyStyle.Render(label), value)
}

// PrintCover prints cover art info with muted label style
func PrintCover(info string) {
	fmt.Printf("  %s %s\n", KeyStyle.Render("Cover:"), info)
}
