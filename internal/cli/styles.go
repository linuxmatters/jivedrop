package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - matching Jivefire
var (
	primaryColor   = lipgloss.Color("#A40000") // Jivedrop red
	accentColor    = lipgloss.Color("#FFA500") // Orange/gold
	successColor   = lipgloss.Color("#00AA00") // Green
	mutedColor     = lipgloss.Color("#888888") // Gray
	highlightColor = lipgloss.Color("#FFFF00") // Yellow
	textColor      = lipgloss.Color("#FFFFFF") // White
)

// Styles
var (
	// Title style - bold red with disco ball emoji
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle style - muted gray
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Section header style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginTop(1).
			MarginBottom(1)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

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
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)
)

// PrintBanner prints the application banner
func PrintBanner() {
	banner := TitleStyle.Render("Jivedrop 🪩")
	subtitle := SubtitleStyle.Render("MP3 Encoder - Drop your podcast audio into RSS-ready MP3s")
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
	fmt.Printf("%s %s\n", HighlightStyle.Render("Warning:"), message)
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
