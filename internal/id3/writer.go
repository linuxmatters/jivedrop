package id3

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
)

// TagInfo holds all the information needed to write ID3v2 tags
type TagInfo struct {
	EpisodeNumber string
	Title         string
	Date          string // Format: "YYYY-MM"
	CoverArtPath  string
}

// WriteTags writes ID3v2.4 tags to an MP3 file
func WriteTags(mp3Path string, info TagInfo) error {
	// Open the MP3 file with ID3v2.4
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: false})
	if err != nil {
		return fmt.Errorf("failed to open MP3 for tagging: %w", err)
	}
	defer tag.Close()

	// Set version to ID3v2.4
	tag.SetVersion(4)

	// TIT2: Title = "Episode: Title"
	titleFrame := fmt.Sprintf("%s: %s", info.EpisodeNumber, info.Title)
	tag.SetTitle(titleFrame)

	// TALB: Album = "Linux Matters"
	tag.SetAlbum("Linux Matters")

	// TRCK: Track number
	tag.AddTextFrame(tag.CommonID("Track number/Position in set"), tag.DefaultEncoding(), info.EpisodeNumber)

	// TPE1: Artist = "Linux Matters"
	tag.SetArtist("Linux Matters")

	// TDRC: Recording date (year and month)
	if info.Date != "" {
		tag.AddTextFrame(tag.CommonID("Recording time"), tag.DefaultEncoding(), info.Date)
	}

	// COMM: Comment = website URL
	commentFrame := id3v2.CommentFrame{
		Encoding:    id3v2.EncodingUTF8,
		Language:    "eng",
		Description: "",
		Text:        "https://linuxmatters.sh/",
	}
	tag.AddCommentFrame(commentFrame)

	// APIC: Cover art
	if info.CoverArtPath != "" {
		if err := addCoverArt(tag, info.CoverArtPath); err != nil {
			return fmt.Errorf("failed to add cover art: %w", err)
		}
	}

	// Save the tag
	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save ID3 tags: %w", err)
	}

	return nil
}

// addCoverArt adds cover artwork as an APIC frame
func addCoverArt(tag *id3v2.Tag, coverPath string) error {
	// Read the cover art file
	artwork, err := os.ReadFile(coverPath)
	if err != nil {
		return fmt.Errorf("failed to read cover art: %w", err)
	}

	// Create APIC frame
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/png",
		PictureType: id3v2.PTFrontCover,
		Description: "Linux Matters Logo",
		Picture:     artwork,
	}

	tag.AddAttachedPicture(pic)

	return nil
}
