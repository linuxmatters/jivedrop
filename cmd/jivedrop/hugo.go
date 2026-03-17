package main

import (
	"fmt"
	"os"

	"github.com/linuxmatters/jivedrop/internal/cli"
	"github.com/linuxmatters/jivedrop/internal/encoder"
	"github.com/linuxmatters/jivedrop/internal/id3"
)

// HugoWorkflow implements the Workflow interface for Hugo mode.
// It reads metadata from Hugo frontmatter and supports frontmatter updates after encoding.
type HugoWorkflow struct {
	// hugoMetadata is set during CollectMetadata and read during PostEncode
	hugoMetadata *encoder.EpisodeMetadata
}

// Validate checks Hugo-specific arguments and file existence.
func (h *HugoWorkflow) Validate() error {
	// Validate markdown argument
	if err := validateHugoMode(); err != nil {
		return err
	}

	// Validate audio file exists
	if _, err := os.Stat(CLI.AudioFile); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", CLI.AudioFile)
	}

	// Validate episode markdown file exists
	if _, err := os.Stat(CLI.EpisodeMD); os.IsNotExist(err) {
		return fmt.Errorf("episode file not found: %s", CLI.EpisodeMD)
	}

	// Validate custom cover art exists (if specified)
	if CLI.Cover != "" {
		if _, err := os.Stat(CLI.Cover); os.IsNotExist(err) {
			return fmt.Errorf("cover art not found: %s", CLI.Cover)
		}
	}

	return nil
}

// CollectMetadata parses Hugo frontmatter, applies defaults and flag overrides,
// and resolves the cover art path.
func (h *HugoWorkflow) CollectMetadata() (id3.TagInfo, string, error) {
	// Parse episode metadata from markdown
	metadata, err := encoder.ParseEpisodeMetadata(CLI.EpisodeMD)
	if err != nil {
		return id3.TagInfo{}, "", fmt.Errorf("failed to parse episode metadata: %w", err)
	}
	h.hugoMetadata = metadata

	// Apply Hugo defaults
	episodeNum := metadata.Episode
	episodeTitle := metadata.Title
	artist := HugoDefaultArtist
	comment := HugoDefaultComment
	date := encoder.FormatDateForID3(metadata.Date)
	var album string

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
	var coverArtPath string
	if CLI.Cover != "" {
		coverArtPath = CLI.Cover
	} else {
		coverArtPath, err = encoder.ResolveCoverArtPath(CLI.EpisodeMD, metadata.EpisodeImage)
		if err != nil {
			return id3.TagInfo{}, "", fmt.Errorf("failed to resolve cover art: %w", err)
		}
	}

	tagInfo := id3.TagInfo{
		EpisodeNumber: episodeNum,
		Title:         episodeTitle,
		Artist:        artist,
		Album:         album,
		Date:          date,
		Comment:       comment,
	}

	return tagInfo, coverArtPath, nil
}

// PostEncode displays podcast statistics and handles frontmatter comparison and update prompting.
func (h *HugoWorkflow) PostEncode(stats *encoder.FileStats, outputPath string) error {
	// Display podcast statistics
	fmt.Println("\nPodcast statistics:")
	cli.PrintLabelValue("•   podcast_duration:", stats.DurationString)
	cli.PrintLabelValue("•   podcast_bytes:", fmt.Sprintf("%d", stats.FileSizeBytes))

	// Check if values differ from existing frontmatter
	needsUpdate := false
	if h.hugoMetadata.PodcastDuration != "" && h.hugoMetadata.PodcastDuration != stats.DurationString {
		cli.PrintWarning(fmt.Sprintf("Duration mismatch: frontmatter has %s, calculated %s",
			h.hugoMetadata.PodcastDuration, stats.DurationString))
		needsUpdate = true
	}
	if h.hugoMetadata.PodcastBytes > 0 && h.hugoMetadata.PodcastBytes != stats.FileSizeBytes {
		cli.PrintWarning(fmt.Sprintf("File size mismatch: frontmatter has %d, calculated %d",
			h.hugoMetadata.PodcastBytes, stats.FileSizeBytes))
		needsUpdate = true
	}

	// Prompt user to update frontmatter if values differ or are missing
	if needsUpdate {
		promptAndUpdateFrontmatter(CLI.EpisodeMD, "\nUpdate frontmatter with new values? [y/N]: ", stats.DurationString, stats.FileSizeBytes)
	} else if h.hugoMetadata.PodcastDuration == "" || h.hugoMetadata.PodcastBytes == 0 {
		promptAndUpdateFrontmatter(CLI.EpisodeMD, "\nAdd podcast_duration and podcast_bytes to frontmatter? [y/N]: ", stats.DurationString, stats.FileSizeBytes)
	}

	return nil
}

// Ensure HugoWorkflow implements Workflow at compile time.
var _ Workflow = (*HugoWorkflow)(nil)
