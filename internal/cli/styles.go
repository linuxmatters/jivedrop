package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Title style - bold blue with disco ball emoji
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	// Subtitle style - muted slate gray
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	// Section header style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			MarginTop(1).
			MarginBottom(1)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SuccessColor)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ErrorColor)

	// Highlight style for important values
	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(HighlightColor)

	// Key-value pair styles
	KeyStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	ValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextColor)

	// Box style for framed content
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
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
	fmt.Printf("%s %s\n", lipgloss.NewStyle().Bold(true).Foreground(SecondaryColor).Render("Warning:"), message)
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
