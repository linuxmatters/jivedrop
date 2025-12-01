package encoder

import (
	"fmt"
	"os"
	"time"

	"github.com/linuxmatters/ffmpeg-statigo"
)

// FileStats holds podcast frontmatter statistics
type FileStats struct {
	DurationString string // HH:MM:SS format
	DurationSecs   int64  // Duration in seconds
	FileSizeBytes  int64  // File size in bytes
}

// GetFileStats extracts duration and file size from an encoded MP3 file
func GetFileStats(mp3Path string) (*FileStats, error) {
	// Get file size
	fileInfo, err := os.Stat(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Open the MP3 file with ffmpeg to get duration
	duration, err := getMP3Duration(mp3Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// Format duration as HH:MM:SS
	durationStr := formatDurationHMS(duration)

	return &FileStats{
		DurationString: durationStr,
		DurationSecs:   duration,
		FileSizeBytes:  fileInfo.Size(),
	}, nil
}

// getMP3Duration opens an MP3 file and extracts its duration
func getMP3Duration(mp3Path string) (int64, error) {
	// Suppress FFmpeg logs
	ffmpeg.AVLogSetLevel(ffmpeg.AVLogError)

	// Open input file
	var fmtCtx *ffmpeg.AVFormatContext
	urlPtr := ffmpeg.ToCStr(mp3Path)
	defer urlPtr.Free()

	if _, err := ffmpeg.AVFormatOpenInput(&fmtCtx, urlPtr, nil, nil); err != nil {
		return 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer ffmpeg.AVFormatCloseInput(&fmtCtx)

	// Find stream information
	if _, err := ffmpeg.AVFormatFindStreamInfo(fmtCtx, nil); err != nil {
		return 0, fmt.Errorf("cannot find stream information: %w", err)
	}

	// Get duration from format context (in AV_TIME_BASE units)
	duration := fmtCtx.Duration()
	if duration <= 0 {
		return 0, fmt.Errorf("invalid duration: %d", duration)
	}

	// Convert from AV_TIME_BASE to seconds
	durationSecs := duration / ffmpeg.AVTimeBase

	return durationSecs, nil
}

// formatDurationHMS converts seconds to HH:MM:SS format
func formatDurationHMS(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
