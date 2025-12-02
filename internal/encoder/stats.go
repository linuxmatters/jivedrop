package encoder

import (
	"fmt"
	"os"
	"time"
)

// FileStats holds podcast frontmatter statistics
type FileStats struct {
	DurationString string // HH:MM:SS format
	DurationSecs   int64  // Duration in seconds
	FileSizeBytes  int64  // File size in bytes
}

// GetFileStats returns file statistics using a pre-calculated duration.
// This avoids re-opening the MP3 file with FFmpeg to extract duration.
func GetFileStats(mp3Path string, durationSecs int64) (*FileStats, error) {
	// Get file size
	fileInfo, err := os.Stat(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Format duration as HH:MM:SS
	durationStr := formatDurationHMS(durationSecs)

	return &FileStats{
		DurationString: durationStr,
		DurationSecs:   durationSecs,
		FileSizeBytes:  fileInfo.Size(),
	}, nil
}

// formatDurationHMS converts seconds to HH:MM:SS format
func formatDurationHMS(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
