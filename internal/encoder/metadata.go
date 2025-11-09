package encoder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EpisodeMetadata holds parsed episode information from Hugo frontmatter
type EpisodeMetadata struct {
	Episode      string    `yaml:"episode"`
	Title        string    `yaml:"title"`
	Date         time.Time `yaml:"Date"`
	EpisodeImage string    `yaml:"episode_image"`
}

// ParseEpisodeMetadata extracts metadata from a Hugo markdown file
func ParseEpisodeMetadata(markdownPath string) (*EpisodeMetadata, error) {
	// Read the file
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read episode file: %w", err)
	}

	// Extract frontmatter between --- delimiters
	frontmatter, err := extractFrontmatter(string(content))
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var meta EpisodeMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Validate required fields
	if meta.Episode == "" {
		return nil, fmt.Errorf("missing required field: episode")
	}
	if meta.Title == "" {
		return nil, fmt.Errorf("missing required field: title")
	}
	if meta.EpisodeImage == "" {
		return nil, fmt.Errorf("missing required field: episode_image")
	}

	return &meta, nil
}

// extractFrontmatter extracts YAML content between --- delimiters
func extractFrontmatter(content string) (string, error) {
	// Find the frontmatter section
	lines := strings.Split(content, "\n")

	var start, end int
	delimiterCount := 0

	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			delimiterCount++
			if delimiterCount == 1 {
				start = i + 1
			} else if delimiterCount == 2 {
				end = i
				break
			}
		}
	}

	if delimiterCount != 2 {
		return "", fmt.Errorf("invalid frontmatter: expected two '---' delimiters, found %d", delimiterCount)
	}

	return strings.Join(lines[start:end], "\n"), nil
}

// ResolveCoverArtPath resolves the episode_image path to an absolute path
// The episode_image in frontmatter is relative to the markdown file
func ResolveCoverArtPath(markdownPath, episodeImage string) (string, error) {
	// Get the directory containing the markdown file
	markdownDir := filepath.Dir(markdownPath)

	// If episodeImage starts with "./", it's relative to markdown location
	if strings.HasPrefix(episodeImage, "./") {
		coverPath := filepath.Join(markdownDir, episodeImage[2:])
		coverPath, err := filepath.Abs(coverPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve cover art path: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(coverPath); err != nil {
			return "", fmt.Errorf("cover art not found: %s", coverPath)
		}

		return coverPath, nil
	}

	// Otherwise, assume it's relative to website root
	// Walk up from markdown to find project root (contains "static" directory)
	projectRoot, err := findProjectRoot(markdownDir)
	if err != nil {
		return "", err
	}

	// Resolve relative to static directory
	coverPath := filepath.Join(projectRoot, "static", strings.TrimPrefix(episodeImage, "/"))
	coverPath, err = filepath.Abs(coverPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cover art path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(coverPath); err != nil {
		return "", fmt.Errorf("cover art not found: %s", coverPath)
	}

	return coverPath, nil
}

// findProjectRoot walks up the directory tree to find the Hugo project root
// (directory containing "static" folder)
func findProjectRoot(startPath string) (string, error) {
	currentPath := startPath

	for {
		// Check if static directory exists here
		staticPath := filepath.Join(currentPath, "static")
		if info, err := os.Stat(staticPath); err == nil && info.IsDir() {
			return currentPath, nil
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)

		// Check if we've reached the root
		if parentPath == currentPath {
			return "", fmt.Errorf("could not find Hugo project root (no 'static' directory found)")
		}

		currentPath = parentPath
	}
}

// FormatDateForID3 formats a time.Time to "YYYY-MM" format for ID3 TDRC tag
func FormatDateForID3(t time.Time) string {
	return t.Format("2006-01")
}
