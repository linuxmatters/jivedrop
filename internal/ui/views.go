package ui

import (
	"fmt"
	"strings"

	"github.com/linuxmatters/jivedrop/internal/encoder"
)

// progressView renders the encoding progress UI
func progressView(m *EncodeModel) string {
	var b strings.Builder

	// Title
	b.WriteString(headerStyle.Render("Encoding to MP3..."))
	b.WriteString("\n\n")

	// Progress bar using bubbles/progress with gradient
	progress := m.calculateProgress() / 100.0 // Convert to 0.0-1.0 range
	b.WriteString(m.progressBar.ViewAs(progress))
	b.WriteString(fmt.Sprintf("  %s", highlightStyle.Render(fmt.Sprintf("%3.0f%%", progress*100))))
	b.WriteString("\n\n")

	// Time and speed info
	elapsed := formatDurationHuman(m.lastUpdateTime.Sub(m.startTime))
	remaining := formatDurationHuman(m.calculateTimeRemaining())
	speed := m.calculateSpeed()

	// Build a visually delightful stats line
	stats := fmt.Sprintf("%s %s   %s %s   %s",
		keyStyle.Render("Elapsed:"),
		highlightStyle.Render(elapsed),
		mutedStyle.Render("~Remaining:"),
		mutedStyle.Render(remaining),
		accentStyle.Render(fmt.Sprintf("%.1fx realtime", speed)),
	)

	b.WriteString(stats)
	b.WriteString("\n\n")

	// Audio specs
	inputSpec := fmt.Sprintf("%s %.1fkHz %s",
		m.inputFormat,
		float64(m.inputRate)/1000.0,
		encoder.FormatChannelMode(m.inputChannels),
	)

	outputSpec := fmt.Sprintf("MP3 %.1fkHz %s CBR %dkbps",
		44.1,
		m.outputMode,
		m.outputBitrate,
	)

	b.WriteString(fmt.Sprintf("%s  %s\n",
		keyStyle.Render("Input:"),
		valueStyle.Render(inputSpec),
	))
	b.WriteString(fmt.Sprintf("%s  %s\n",
		keyStyle.Render("Output:"),
		valueStyle.Render(outputSpec),
	))

	return boxStyle.Render(b.String())
}

// completeView renders the completion message
func completeView(m *EncodeModel) string {
	elapsed := formatDurationHuman(m.lastUpdateTime.Sub(m.startTime))
	speed := m.calculateSpeed()

	msg := fmt.Sprintf("%s Encoded in %s (%.1fx realtime)",
		successStyle.Render("✓"),
		valueStyle.Render(elapsed),
		speed,
	)

	return msg + "\n"
}

// errorView renders an error message
func errorView(err error) string {
	return fmt.Sprintf("%s %s\n",
		errorStyle.Render("Error:"),
		err.Error(),
	)
}
