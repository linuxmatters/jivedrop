package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/linuxmatters/jivedrop/internal/cli"
	"github.com/linuxmatters/jivedrop/internal/encoder"
	"github.com/linuxmatters/jivedrop/internal/id3"
	"github.com/linuxmatters/jivedrop/internal/ui"
)

// version is set via ldflags at build time
// Local dev builds: "dev"
// Release builds: git tag (e.g. "v0.1.0")
var version = "dev"

// coverArtResult holds the outcome of cover art processing
type coverArtResult struct {
	data []byte
	err  error
}

type WorkflowMode int

const (
	HugoMode WorkflowMode = iota
	StandaloneMode
)

// Hugo mode defaults for Linux Matters podcast
const (
	HugoDefaultArtist  = "Linux Matters"
	HugoDefaultComment = "https://linuxmatters.sh"
	HugoDefaultPrefix  = "LMP"
)

var CLI struct {
	AudioFile string `arg:"" name:"audio-file" help:"Path to audio file (WAV, FLAC)" optional:""`
	EpisodeMD string `arg:"" name:"episode-md" help:"Path to episode markdown file (Hugo mode)" optional:""`

	// Metadata flags (standalone mode or Hugo overrides)
	Num        string `help:"Episode number"`
	Title      string `help:"Episode title"`
	Artist     string `help:"Artist name (defaults to 'Linux Matters' in Hugo mode)"`
	Album      string `help:"Album name (defaults to artist value if omitted)"`
	Date       string `help:"Release date (YYYY-MM-DD format)"`
	Comment    string `help:"Comment URL (defaults to 'https://linuxmatters.sh' in Hugo mode)"`
	Cover      string `help:"Cover art path"`
	OutputPath string `help:"Output file or directory path"`

	// Encoding options
	Stereo  bool `help:"Encode as stereo at 192kbps (default: mono at 112kbps)"`
	Version bool `help:"Show version information"`
}

// detectMode determines if this is Hugo or Standalone workflow
func detectMode() WorkflowMode {
	// If AudioFile is empty, we have no arguments - show help
	if CLI.AudioFile == "" {
		return HugoMode // Return value doesn't matter, we'll exit in main
	}

	// If second argument exists and is a .md file, we're in Hugo mode
	if CLI.EpisodeMD != "" && strings.HasSuffix(strings.ToLower(CLI.EpisodeMD), ".md") {
		return HugoMode
	}

	return StandaloneMode
}

// sanitiseForFilename replaces spaces and invalid characters for safe filenames
func sanitiseForFilename(s string) string {
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Convert to lowercase for consistency
	s = strings.ToLower(s)
	// Remove any characters that aren't alphanumeric, hyphens, or underscores
	// Keep dots for file extensions
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return -1
	}, s)
}

// generateFilename creates the output filename based on mode and metadata
func generateFilename(mode WorkflowMode, num, artist string) string {
	if mode == HugoMode {
		// Hugo mode: LMP{num}.mp3 unless artist is overridden
		if CLI.Artist != "" && CLI.Artist != HugoDefaultArtist {
			// Custom artist provided
			sanitisedArtist := sanitiseForFilename(artist)
			return fmt.Sprintf("%s-%s.mp3", sanitisedArtist, num)
		}
		// Default Linux Matters format
		return fmt.Sprintf("%s%s.mp3", HugoDefaultPrefix, num)
	}

	// Standalone mode: {artist}-{num}.mp3 or episode-{num}.mp3 fallback
	if artist != "" {
		sanitisedArtist := sanitiseForFilename(artist)
		return fmt.Sprintf("%s-%s.mp3", sanitisedArtist, num)
	}

	return fmt.Sprintf("episode-%s.mp3", num)
}

// resolveOutputPath determines final output file path
func resolveOutputPath(mode WorkflowMode, num, artist string) (string, error) {
	if CLI.OutputPath == "" {
		// No output path specified, use current directory with generated filename
		filename := generateFilename(mode, num, artist)
		return filename, nil
	}

	// Check if OutputPath is a directory or file
	stat, err := os.Stat(CLI.OutputPath)
	if err == nil {
		if stat.IsDir() {
			// It's a directory, generate filename
			filename := generateFilename(mode, num, artist)
			return filepath.Join(CLI.OutputPath, filename), nil
		}
		// It's a file path, use as-is
		return CLI.OutputPath, nil
	}

	// Path doesn't exist - check if it ends with / to determine intent
	if strings.HasSuffix(CLI.OutputPath, "/") {
		return "", fmt.Errorf("output directory does not exist: %s", CLI.OutputPath)
	}

	// Assume it's a file path (may be in non-existent directory)
	dir := filepath.Dir(CLI.OutputPath)
	if dir != "." && dir != "" {
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			return "", fmt.Errorf("output directory does not exist: %s", dir)
		}
	}

	return CLI.OutputPath, nil
}

// promptAndUpdateFrontmatter prompts the user and updates the frontmatter with podcast stats
func promptAndUpdateFrontmatter(markdownPath, promptMsg, duration string, bytes int64) {
	fmt.Print(promptMsg)
	var response string
	_, _ = fmt.Scanln(&response)

	if strings.ToLower(strings.TrimSpace(response)) == "y" {
		if err := encoder.UpdateFrontmatter(markdownPath, duration, bytes); err != nil {
			cli.PrintError(fmt.Sprintf("Failed to update frontmatter: %v", err))
		} else {
			cli.PrintSuccess("Frontmatter updated successfully")
		}
	} else {
		cli.PrintInfo("Frontmatter not updated")
	}
}

// encode runs the full encoding pipeline: display info, create encoder, run
// Bubbletea UI, process cover art concurrently, write ID3 tags, and extract
// file statistics. Returns nil stats (with nil error) when stats extraction
// fails but the MP3 was written successfully.
func encode(mode WorkflowMode, tagInfo id3.TagInfo, coverArtPath, outputPath string) (*encoder.FileStats, error) {
	// Display encoding info
	cli.PrintSuccessLabel("Ready to encode:", fmt.Sprintf("%s -> MP3", CLI.AudioFile))
	cli.PrintLabelValue("• Episode:", fmt.Sprintf("%s - %s", tagInfo.EpisodeNumber, tagInfo.Title))
	if mode == HugoMode {
		cli.PrintLabelValue("• Episode markdown:", CLI.EpisodeMD)
	}
	cli.PrintLabelValue("• Output:", outputPath)
	if CLI.Stereo {
		cli.PrintLabelValue("• Encoding mode:", "Stereo 192kbps")
	} else {
		cli.PrintLabelValue("• Encoding mode:", "Mono 112kbps")
	}

	// Create encoder
	enc, err := encoder.New(encoder.Config{
		InputPath:  CLI.AudioFile,
		OutputPath: outputPath,
		Stereo:     CLI.Stereo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}
	defer enc.Close()

	// Initialize encoder
	if err := enc.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize encoder: %w", err)
	}

	// Get input info
	sampleRate, channels, format := enc.GetInputInfo()
	channelMode := encoder.FormatChannelMode(channels)
	cli.PrintLabelValue("• Input:", fmt.Sprintf("%s %dHz %s", format, sampleRate, channelMode))

	// Determine output bitrate and mode
	outputBitrate := 112
	outputMode := "mono"
	if CLI.Stereo {
		outputBitrate = 192
		outputMode = "stereo"
	}

	// Start cover art processing concurrently with encoding
	coverArtChan := make(chan coverArtResult, 1)
	go func() {
		if coverArtPath == "" {
			coverArtChan <- coverArtResult{data: nil, err: nil}
			return
		}

		artwork, artErr := id3.ScaleCoverArt(coverArtPath, nil)
		coverArtChan <- coverArtResult{data: artwork, err: artErr}
	}()

	// Start encoding with Bubbletea UI
	fmt.Println()
	encodeModel := ui.NewEncodeModel(enc, outputMode, outputBitrate)

	p := tea.NewProgram(encodeModel)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("UI error: %w", err)
	}

	// Check for encoding errors
	if encModel, ok := finalModel.(*ui.EncodeModel); ok {
		if encModel.Error() != nil {
			// Clean up partial output file
			os.Remove(outputPath)
			return nil, fmt.Errorf("encoding failed: %w", encModel.Error())
		}
	}

	// Collect cover art result from concurrent processing
	coverResult := <-coverArtChan
	if coverResult.err != nil {
		cli.PrintInfo(fmt.Sprintf("MP3 file created but missing cover art: %s", outputPath))
		return nil, fmt.Errorf("failed to process cover art: %w", coverResult.err)
	}

	// Write ID3v2 tags
	fmt.Println("\nEmbedding ID3v2 tags...")
	tagInfo.CoverArtData = coverResult.data

	if err := id3.WriteTags(outputPath, tagInfo); err != nil {
		cli.PrintInfo(fmt.Sprintf("MP3 file created but missing metadata: %s", outputPath))
		return nil, fmt.Errorf("failed to write ID3 tags: %w", err)
	}

	cli.PrintSuccessLabel("Complete:", outputPath)

	// Extract file statistics using duration from encoder (avoids re-opening file)
	durationSecs := enc.GetDurationSecs()
	stats, err := encoder.GetFileStats(outputPath, durationSecs)
	if err != nil {
		cli.PrintWarning(fmt.Sprintf("Could not extract file statistics: %v", err))
		return nil, nil
	}

	return stats, nil
}

func main() {
	os.Exit(run())
}

func run() int {
	ctx := kong.Parse(&CLI,
		kong.Name("jivedrop"),
		kong.Description("Drop the mix, ship the show—metadata, cover art, and all."),
		kong.Vars{"version": version},
		kong.UsageOnError(),
		kong.Help(cli.StyledHelpPrinter(kong.HelpOptions{Compact: true})),
	)

	// Handle version flag
	if CLI.Version {
		cli.PrintVersion(version)
		return 0
	}

	// If no audio file provided, show help
	if CLI.AudioFile == "" {
		_ = ctx.PrintUsage(false)
		return 0
	}

	// Detect workflow mode, validate, and collect metadata
	mode := detectMode()
	wf := newWorkflow(mode)

	if err := wf.Validate(); err != nil {
		cli.PrintError(err.Error())
		return 1
	}

	tagInfo, coverArtPath, err := wf.CollectMetadata()
	if err != nil {
		cli.PrintError(err.Error())
		return 1
	}

	// Resolve output path
	outputPath, err := resolveOutputPath(mode, tagInfo.EpisodeNumber, tagInfo.Artist)
	if err != nil {
		cli.PrintError(fmt.Sprintf("Failed to resolve output path: %v", err))
		return 1
	}

	// Encode audio, process cover art, write ID3 tags, and collect file statistics
	stats, err := encode(mode, tagInfo, coverArtPath, outputPath)
	if err != nil {
		cli.PrintError(err.Error())
		return 1
	}

	// Stats may be nil when extraction failed but encoding succeeded
	if stats == nil {
		return 0
	}

	// Post-encode: display stats and handle mode-specific operations
	if err := wf.PostEncode(stats, outputPath); err != nil {
		cli.PrintError(err.Error())
		return 1
	}

	return 0
}
