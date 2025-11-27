package id3

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func TestWriteTags(t *testing.T) {
	// Create a minimal valid MP3 file for testing
	tmpDir := t.TempDir()
	mp3Path := filepath.Join(tmpDir, "test.mp3")

	// Copy test MP3 from testdata
	// The MP3 is created by encoder integration tests if it doesn't exist
	testMP3 := "../../testdata/LMP0.mp3"
	input, err := os.ReadFile(testMP3)
	if err != nil {
		t.Skipf("Test MP3 not found at %s: %v", testMP3, err)
	}
	if err := os.WriteFile(mp3Path, input, 0644); err != nil {
		t.Fatalf("Failed to create test MP3: %v", err)
	}

	// Test tag info
	info := TagInfo{
		EpisodeNumber: "67",
		Title:         "Mirrors, Motors and Makefiles",
		Artist:        "Linux Matters",
		Album:         "Linux Matters",
		Date:          "2025-11",
		Comment:       "https://linuxmatters.sh/",
		CoverArtPath:  "", // Skip cover art for basic test
	}

	// Write tags
	err = WriteTags(mp3Path, info)
	if err != nil {
		t.Fatalf("WriteTags failed: %v", err)
	}

	// Verify tags were written correctly
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("Failed to open MP3 for verification: %v", err)
	}
	defer tag.Close()

	// Verify TIT2 (Title)
	expectedTitle := "67: Mirrors, Motors and Makefiles"
	if tag.Title() != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, tag.Title())
	}

	// Verify TALB (Album)
	if tag.Album() != "Linux Matters" {
		t.Errorf("Expected album 'Linux Matters', got '%s'", tag.Album())
	}

	// Verify TPE1 (Artist)
	if tag.Artist() != "Linux Matters" {
		t.Errorf("Expected artist 'Linux Matters', got '%s'", tag.Artist())
	}

	// Verify COMM (Comment)
	comments := tag.GetFrames(tag.CommonID("Comments"))
	if len(comments) == 0 {
		t.Error("Expected comment frame, got none")
	} else {
		commentFrame, ok := comments[0].(id3v2.CommentFrame)
		if !ok {
			t.Error("Comment frame is not of type CommentFrame")
		} else if commentFrame.Text != "https://linuxmatters.sh/" {
			t.Errorf("Expected comment 'https://linuxmatters.sh/', got '%s'", commentFrame.Text)
		}
	}
}

func TestWriteTags_NonExistentFile(t *testing.T) {
	info := TagInfo{
		EpisodeNumber: "67",
		Title:         "Test Episode",
		Date:          "2025-11",
	}

	err := WriteTags("/nonexistent/path/test.mp3", info)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestWriteTags_WithDate(t *testing.T) {
	tmpDir := t.TempDir()
	mp3Path := filepath.Join(tmpDir, "test.mp3")

	// Copy test MP3
	testMP3 := "../../testdata/LMP0.mp3"
	input, err := os.ReadFile(testMP3)
	if err != nil {
		t.Skipf("Test MP3 not found: %v", err)
	}
	if err := os.WriteFile(mp3Path, input, 0644); err != nil {
		t.Fatalf("Failed to create test MP3: %v", err)
	}

	info := TagInfo{
		EpisodeNumber: "67",
		Title:         "Test Episode",
		Date:          "2025-11",
	}

	if err := WriteTags(mp3Path, info); err != nil {
		t.Fatalf("WriteTags failed: %v", err)
	}

	// Verify date tag
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("Failed to open MP3: %v", err)
	}
	defer tag.Close()

	dateFrames := tag.GetFrames(tag.CommonID("Recording time"))
	if len(dateFrames) == 0 {
		t.Error("Expected recording time frame, got none")
	}
}
