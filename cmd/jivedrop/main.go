package main

import (
	"fmt"
	"os"
	"path/filepath"

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

var CLI struct {
	EpisodeMD string `arg:"" name:"episode-md" help:"Path to episode markdown file (e.g., 67.md)" optional:""`
	InputFile string `arg:"" name:"input-file" help:"Path to audio file (WAV, FLAC)" optional:""`
	OutputDir string `arg:"" name:"output-dir" help:"Output directory (default: current directory)" optional:""`
	Stereo    bool   `help:"Encode as stereo at 192kbps (default: mono at 112kbps)"`
	Cover     string `help:"Custom cover art path (overrides frontmatter)"`
	Version   bool   `help:"Show version information"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("jivedrop"),
		kong.Description("Drop your podcast audio into RSS-ready MP3s with embedded artwork and ID3 metadata."),
		kong.Vars{"version": version},
		kong.UsageOnError(),
		kong.Help(cli.StyledHelpPrinter(kong.HelpOptions{Compact: true})),
	)

	// Handle version flag
	if CLI.Version {
		cli.PrintVersion(version)
		os.Exit(0)
	}

	// Validate required arguments when not showing version
	if CLI.EpisodeMD == "" || CLI.InputFile == "" {
		cli.PrintError("<episode-md> and <input-file> are required")
		os.Exit(1)
	}

	_ = ctx // Kong context available for future use

	// Validate episode markdown file exists
	if _, err := os.Stat(CLI.EpisodeMD); os.IsNotExist(err) {
		cli.PrintError(fmt.Sprintf("Episode file not found: %s", CLI.EpisodeMD))
		cli.PrintInfo("Make sure the episode markdown file exists.")
		os.Exit(1)
	}

	// Validate input audio file exists
	if _, err := os.Stat(CLI.InputFile); os.IsNotExist(err) {
		cli.PrintError(fmt.Sprintf("Input file not found: %s", CLI.InputFile))
		cli.PrintInfo("Make sure the audio file exists.")
		os.Exit(1)
	}

	// Validate output directory exists (if specified)
	if CLI.OutputDir != "" {
		if stat, err := os.Stat(CLI.OutputDir); os.IsNotExist(err) {
			cli.PrintError(fmt.Sprintf("Output directory not found: %s", CLI.OutputDir))
			cli.PrintInfo("Make sure the directory exists.")
			os.Exit(1)
		} else if !stat.IsDir() {
			cli.PrintError(fmt.Sprintf("Output path is not a directory: %s", CLI.OutputDir))
			os.Exit(1)
		}
	}

	// Validate custom cover art exists (if specified)
	if CLI.Cover != "" {
		if _, err := os.Stat(CLI.Cover); os.IsNotExist(err) {
			cli.PrintError(fmt.Sprintf("Cover art not found: %s", CLI.Cover))
			cli.PrintInfo("Make sure the cover art file exists.")
			os.Exit(1)
		}
	}

	// Parse episode metadata
	metadata, err := encoder.ParseEpisodeMetadata(CLI.EpisodeMD)
	if err != nil {
		cli.PrintError(fmt.Sprintf("Failed to parse episode metadata: %v", err))
		os.Exit(1)
	}

	// Resolve cover art path (use custom cover if provided, otherwise from metadata)
	var coverArtPath string
	if CLI.Cover != "" {
		coverArtPath = CLI.Cover
	} else {
		coverArtPath, err = encoder.ResolveCoverArtPath(CLI.EpisodeMD, metadata.EpisodeImage)
		if err != nil {
			cli.PrintError(fmt.Sprintf("Failed to resolve cover art: %v", err))
			cli.PrintInfo("Use --cover flag to specify a custom cover art path.")
			os.Exit(1)
		}
	}

	cli.PrintSuccess(fmt.Sprintf("Ready to encode: %s -> MP3", CLI.InputFile))
	cli.PrintInfo(fmt.Sprintf("Episode: %s - %s", metadata.Episode, metadata.Title))
	cli.PrintInfo(fmt.Sprintf("Episode markdown: %s", CLI.EpisodeMD))
	if CLI.OutputDir != "" {
		cli.PrintInfo(fmt.Sprintf("Output directory: %s", CLI.OutputDir))
	}
	if CLI.Stereo {
		cli.PrintInfo("Encoding mode: Stereo 192kbps")
	} else {
		cli.PrintInfo("Encoding mode: Mono 112kbps")
	}

	// Determine output path using episode number
	outputDir := CLI.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	outputPath := filepath.Join(outputDir, fmt.Sprintf("LMP%s.mp3", metadata.Episode))

	// Create encoder
	enc, err := encoder.New(encoder.Config{
		InputPath:  CLI.InputFile,
		OutputPath: outputPath,
		Stereo:     CLI.Stereo,
	})
	if err != nil {
		cli.PrintError(fmt.Sprintf("Failed to create encoder: %v", err))
		os.Exit(1)
	}
	defer enc.Close()

	// Initialize encoder
	if err := enc.Initialize(); err != nil {
		cli.PrintError(fmt.Sprintf("Failed to initialize encoder: %v", err))
		os.Exit(1)
	}

	// Get input info
	sampleRate, channels, format := enc.GetInputInfo()
	channelMode := "mono"
	if channels == 2 {
		channelMode = "stereo"
	} else if channels > 2 {
		channelMode = fmt.Sprintf("%dch", channels)
	}
	cli.PrintInfo(fmt.Sprintf("Input: %s %dHz %s", format, sampleRate, channelMode))

	// Determine output bitrate and mode
	outputBitrate := 112
	outputMode := "mono"
	if CLI.Stereo {
		outputBitrate = 192
		outputMode = "stereo"
	}

	// Start encoding with Bubbletea UI
	fmt.Println()
	encodeModel := ui.NewEncodeModel(enc, outputMode, outputBitrate)

	p := tea.NewProgram(encodeModel)
	finalModel, err := p.Run()
	if err != nil {
		cli.PrintError(fmt.Sprintf("UI error: %v", err))
		os.Exit(1)
	}

	// Check for encoding errors
	if encModel, ok := finalModel.(*ui.EncodeModel); ok {
		if encModel.Error() != nil {
			cli.PrintError(fmt.Sprintf("Encoding failed: %v", encModel.Error()))
			// Clean up partial output file
			os.Remove(outputPath)
			os.Exit(1)
		}
	}

	// Write ID3v2 tags
	fmt.Println("\nEmbedding ID3v2 tags...")
	tagInfo := id3.TagInfo{
		EpisodeNumber: metadata.Episode,
		Title:         metadata.Title,
		Date:          encoder.FormatDateForID3(metadata.Date),
		CoverArtPath:  coverArtPath,
	}

	if err := id3.WriteTags(outputPath, tagInfo); err != nil {
		cli.PrintError(fmt.Sprintf("Failed to write ID3 tags: %v", err))
		cli.PrintInfo(fmt.Sprintf("MP3 file created but missing metadata: %s", outputPath))
		os.Exit(1)
	}

	cli.PrintSuccess(fmt.Sprintf("Complete: %s", outputPath))
}
