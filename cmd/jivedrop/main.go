package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/linuxmatters/jivedrop/internal/cli"
	"github.com/linuxmatters/jivedrop/internal/encoder"
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

	cli.PrintSuccess(fmt.Sprintf("Ready to encode: %s -> MP3", CLI.InputFile))
	cli.PrintInfo(fmt.Sprintf("Episode markdown: %s", CLI.EpisodeMD))
	if CLI.OutputDir != "" {
		cli.PrintInfo(fmt.Sprintf("Output directory: %s", CLI.OutputDir))
	}
	if CLI.Stereo {
		cli.PrintInfo("Encoding mode: Stereo 192kbps")
	} else {
		cli.PrintInfo("Encoding mode: Mono 112kbps")
	}
	if CLI.Cover != "" {
		cli.PrintInfo(fmt.Sprintf("Custom cover art: %s", CLI.Cover))
	}

	// Determine output path
	outputDir := CLI.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	outputPath := filepath.Join(outputDir, "output.mp3") // TODO: Use episode number for filename

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

	// Start encoding with progress callback
	fmt.Println("\nEncoding to MP3...")
	startTime := time.Now()

	lastProgress := 0
	err = enc.Encode(func(samplesProcessed, totalSamples int64) {
		if totalSamples > 0 {
			progress := int((samplesProcessed * 100) / totalSamples)
			// Only print every 5% to avoid flooding output
			if progress >= lastProgress+5 || progress == 100 {
				fmt.Printf("  Progress: %d%%\r", progress)
				lastProgress = progress
			}
		}
	})

	if err != nil {
		cli.PrintError(fmt.Sprintf("Encoding failed: %v", err))
		// Clean up partial output file
		os.Remove(outputPath)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n")
	cli.PrintSuccess(fmt.Sprintf("Encoded in %.1fs", elapsed.Seconds()))
	cli.PrintSuccess(fmt.Sprintf("Output: %s", outputPath))
}
