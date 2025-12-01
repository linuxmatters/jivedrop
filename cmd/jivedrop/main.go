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

// validateHugoMode validates Hugo workflow arguments
func validateHugoMode() error {
	// In Hugo mode, episode markdown is required
	if CLI.EpisodeMD == "" {
		return fmt.Errorf("Hugo mode requires episode markdown file as second argument")
	}

	if !strings.HasSuffix(strings.ToLower(CLI.EpisodeMD), ".md") {
		return fmt.Errorf("episode markdown file must have .md extension: %s", CLI.EpisodeMD)
	}

	return nil
}

// validateStandaloneMode validates standalone workflow arguments
func validateStandaloneMode() error {
	// In standalone mode, --title, --num, and --cover are required
	if CLI.Title == "" {
		return fmt.Errorf("standalone mode requires --title flag")
	}

	if CLI.Num == "" {
		return fmt.Errorf("standalone mode requires --num flag (episode number)")
	}

	if CLI.Cover == "" {
		return fmt.Errorf("standalone mode requires --cover flag (cover art path)")
	}

	return nil
}

// sanitiseForFilename replaces spaces and invalid characters for safe filenames
func sanitiseForFilename(s string) string {
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Convert to lowercase for consistency
	s = strings.ToLower(s)
	// Remove any characters that aren't alphanumeric, hyphens, or underscores
	// Keep dots for file extensions
	result := strings.Builder{}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
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
	fmt.Scanln(&response)

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

func main() {
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
		os.Exit(0)
	}

	// If no audio file provided, show help
	if CLI.AudioFile == "" {
		_ = ctx.PrintUsage(false)
		os.Exit(0)
	}

	// Detect workflow mode and validate arguments
	mode := detectMode()

	// Mode-specific validation
	if mode == HugoMode {
		if err := validateHugoMode(); err != nil {
			cli.PrintError(err.Error())
			os.Exit(1)
		}
	} else {
		if err := validateStandaloneMode(); err != nil {
			cli.PrintError(err.Error())
			os.Exit(1)
		}
	}

	_ = ctx // Kong context available for future use

	// Validate audio file exists
	if _, err := os.Stat(CLI.AudioFile); os.IsNotExist(err) {
		cli.PrintError(fmt.Sprintf("Audio file not found: %s", CLI.AudioFile))
		cli.PrintInfo("Make sure the audio file exists.")
		os.Exit(1)
	}

	// Validate episode markdown file exists (Hugo mode)
	if mode == HugoMode {
		if _, err := os.Stat(CLI.EpisodeMD); os.IsNotExist(err) {
			cli.PrintError(fmt.Sprintf("Episode file not found: %s", CLI.EpisodeMD))
			cli.PrintInfo("Make sure the episode markdown file exists.")
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

	// Collect metadata based on workflow mode
	var episodeNum, episodeTitle, artist, album, date, comment, coverArtPath string
	var hugoMetadata *encoder.EpisodeMetadata

	if mode == HugoMode {
		// Parse episode metadata from markdown
		var err error
		hugoMetadata, err = encoder.ParseEpisodeMetadata(CLI.EpisodeMD)
		if err != nil {
			cli.PrintError(fmt.Sprintf("Failed to parse episode metadata: %v", err))
			os.Exit(1)
		}

		// Apply Hugo defaults
		episodeNum = hugoMetadata.Episode
		episodeTitle = hugoMetadata.Title
		artist = HugoDefaultArtist
		comment = HugoDefaultComment
		date = encoder.FormatDateForID3(hugoMetadata.Date)

		// Allow flag overrides
		if CLI.Artist != "" {
			artist = CLI.Artist
		}
		if CLI.Album != "" {
			album = CLI.Album
		} else {
			album = artist // Inherit from artist
		}
		if CLI.Comment != "" {
			comment = CLI.Comment
		}
		if CLI.Title != "" {
			episodeTitle = CLI.Title
		}
		if CLI.Num != "" {
			episodeNum = CLI.Num
		}
		if CLI.Date != "" {
			date = CLI.Date
		}

		// Resolve cover art path
		if CLI.Cover != "" {
			coverArtPath = CLI.Cover
		} else {
			coverArtPath, err = encoder.ResolveCoverArtPath(CLI.EpisodeMD, hugoMetadata.EpisodeImage)
			if err != nil {
				cli.PrintError(fmt.Sprintf("Failed to resolve cover art: %v", err))
				cli.PrintInfo("Use --cover flag to specify a custom cover art path.")
				os.Exit(1)
			}
		}
	} else {
		// Standalone mode: use CLI flags
		episodeNum = CLI.Num
		episodeTitle = CLI.Title
		artist = CLI.Artist
		album = CLI.Album
		date = CLI.Date
		comment = CLI.Comment
		coverArtPath = CLI.Cover

		// Apply defaults for standalone mode
		if album == "" && artist != "" {
			album = artist // Inherit from artist
		}
	}

	// Resolve output path
	outputPath, err := resolveOutputPath(mode, episodeNum, artist)
	if err != nil {
		cli.PrintError(fmt.Sprintf("Failed to resolve output path: %v", err))
		os.Exit(1)
	}

	// Display encoding info
	cli.PrintSuccessLabel("Ready to encode:", fmt.Sprintf("%s -> MP3", CLI.AudioFile))
	cli.PrintLabelValue("• Episode:", fmt.Sprintf("%s - %s", episodeNum, episodeTitle))
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
	cli.PrintLabelValue("• Input:", fmt.Sprintf("%s %dHz %s", format, sampleRate, channelMode))

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
		EpisodeNumber: episodeNum,
		Title:         episodeTitle,
		Artist:        artist,
		Album:         album,
		Date:          date,
		Comment:       comment,
		CoverArtPath:  coverArtPath,
	}

	if err := id3.WriteTags(outputPath, tagInfo); err != nil {
		cli.PrintError(fmt.Sprintf("Failed to write ID3 tags: %v", err))
		cli.PrintInfo(fmt.Sprintf("MP3 file created but missing metadata: %s", outputPath))
		os.Exit(1)
	}

	cli.PrintSuccessLabel("Complete:", outputPath)

	// Extract file statistics
	stats, err := encoder.GetFileStats(outputPath)
	if err != nil {
		cli.PrintWarning(fmt.Sprintf("Could not extract file statistics: %v", err))
		return
	}

	// Display podcast statistics (both modes)
	fmt.Println("\nPodcast statistics:")
	cli.PrintLabelValue("•   podcast_duration:", stats.DurationString)
	cli.PrintLabelValue("•   podcast_bytes:", fmt.Sprintf("%d", stats.FileSizeBytes))

	// Only handle frontmatter updates in Hugo mode
	if mode == StandaloneMode {
		return
	}

	// Hugo mode: check and update frontmatter if needed
	// Check if values differ from existing frontmatter
	needsUpdate := false
	if hugoMetadata.PodcastDuration != "" && hugoMetadata.PodcastDuration != stats.DurationString {
		cli.PrintWarning(fmt.Sprintf("Duration mismatch: frontmatter has %s, calculated %s",
			hugoMetadata.PodcastDuration, stats.DurationString))
		needsUpdate = true
	}
	if hugoMetadata.PodcastBytes > 0 && hugoMetadata.PodcastBytes != stats.FileSizeBytes {
		cli.PrintWarning(fmt.Sprintf("File size mismatch: frontmatter has %d, calculated %d",
			hugoMetadata.PodcastBytes, stats.FileSizeBytes))
		needsUpdate = true
	}

	// Prompt user to update frontmatter if values differ or are missing
	if needsUpdate {
		promptAndUpdateFrontmatter(CLI.EpisodeMD, "\nUpdate frontmatter with new values? [y/N]: ", stats.DurationString, stats.FileSizeBytes)
	} else if hugoMetadata.PodcastDuration == "" || hugoMetadata.PodcastBytes == 0 {
		// If frontmatter is missing these fields, offer to add them
		promptAndUpdateFrontmatter(CLI.EpisodeMD, "\nAdd podcast_duration and podcast_bytes to frontmatter? [y/N]: ", stats.DurationString, stats.FileSizeBytes)
	}
}
