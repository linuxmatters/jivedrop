package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSanitiseForFilename tests filename sanitisation for dangerous and special characters
func TestSanitiseForFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{
			name:     "simple alphanumeric",
			input:    "Linux Matters",
			expected: "linux-matters",
		},
		// Special characters that should be removed
		{
			name:     "forward slash",
			input:    "AC/DC",
			expected: "acdc",
		},
		{
			name:     "backslash",
			input:    "Guns\\Roses",
			expected: "gunsroses",
		},
		{
			name:     "apostrophe",
			input:    "Guns N' Roses",
			expected: "guns-n-roses",
		},
		{
			name:     "ampersand",
			input:    "Tom & Jerry",
			expected: "tom--jerry",
		},
		{
			name:     "asterisk",
			input:    "The *Clash*",
			expected: "the-clash",
		},
		{
			name:     "question mark",
			input:    "Who?",
			expected: "who",
		},
		{
			name:     "exclamation mark",
			input:    "Bang!",
			expected: "bang",
		},
		// Unicode and non-ASCII characters
		{
			name:     "unicode umlauts stripped",
			input:    "Björk",
			expected: "bjrk",
		},
		{
			name:     "accented characters stripped",
			input:    "Café du Monde",
			expected: "caf-du-monde",
		},
		{
			name:     "chinese characters stripped",
			input:    "Podcast 中文 Show",
			expected: "podcast--show",
		},
		// Multiple special characters
		{
			name:     "multiple consecutive special chars",
			input:    "Episode!!!???",
			expected: "episode",
		},
		// Dots and underscores (preserved)
		{
			name:     "dots preserved",
			input:    "Hello...World",
			expected: "hello...world",
		},
		{
			name:     "underscores preserved",
			input:    "Hello_World",
			expected: "hello_world",
		},
		{
			name:     "mixed dots underscores hyphens",
			input:    "Hello.World_Test-Case",
			expected: "hello.world_test-case",
		},
		// Whitespace handling
		{
			name:     "leading and trailing spaces",
			input:    "  Podcast Show  ",
			expected: "--podcast-show--",
		},
		{
			name:     "multiple spaces between words",
			input:    "The   Podcast   Show",
			expected: "the---podcast---show",
		},
		{
			name:     "tabs and mixed whitespace",
			input:    "Hello\tWorld",
			expected: "helloworld",
		},
		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "---",
		},
		{
			name:     "only special characters",
			input:    "!!!???&&&",
			expected: "",
		},
		{
			name:     "numbers preserved",
			input:    "Episode 42",
			expected: "episode-42",
		},
		{
			name:     "mixed case with numbers",
			input:    "PoDCaSt 99 ShOw",
			expected: "podcast-99-show",
		},
		// Real-world examples
		{
			name:     "linux matters real example",
			input:    "Linux Matters",
			expected: "linux-matters",
		},
		{
			name:     "spotify podcast example",
			input:    "The Daily Show with Trevor Noah",
			expected: "the-daily-show-with-trevor-noah",
		},
		{
			name:     "special music artist",
			input:    "U2 / Bono",
			expected: "u2--bono",
		},
		{
			name:     "symbol-heavy artist",
			input:    "$$$ (Money) $$$",
			expected: "-money-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitiseForFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitiseForFilename(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGenerateFilename tests filename generation for both Hugo and Standalone modes
func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name      string
		mode      WorkflowMode
		num       string
		artist    string
		cliArtist string // Simulates CLI.Artist global
		expected  string
	}{
		// Hugo mode - default behaviour
		{
			name:      "hugo default simple",
			mode:      HugoMode,
			num:       "67",
			artist:    "",
			cliArtist: "",
			expected:  "LMP67.mp3",
		},
		{
			name:      "hugo episode 0",
			mode:      HugoMode,
			num:       "0",
			artist:    "",
			cliArtist: "",
			expected:  "LMP0.mp3",
		},
		{
			name:      "hugo large episode number",
			mode:      HugoMode,
			num:       "999",
			artist:    "",
			cliArtist: "",
			expected:  "LMP999.mp3",
		},
		// Hugo mode - custom artist override
		{
			name:      "hugo with custom artist override",
			mode:      HugoMode,
			num:       "67",
			artist:    "Custom Podcast",
			cliArtist: "Custom Podcast",
			expected:  "custom-podcast-67.mp3",
		},
		{
			name:      "hugo with special chars in artist",
			mode:      HugoMode,
			num:       "42",
			artist:    "The (Real) Show",
			cliArtist: "The (Real) Show",
			expected:  "the-real-show-42.mp3",
		},
		// Hugo mode - Linux Matters default not triggered by override
		{
			name:      "hugo with linux matters artist keeps default",
			mode:      HugoMode,
			num:       "50",
			artist:    "Linux Matters",
			cliArtist: "Linux Matters",
			expected:  "LMP50.mp3",
		},
		{
			name:      "hugo empty cli artist keeps default",
			mode:      HugoMode,
			num:       "55",
			artist:    "Other",
			cliArtist: "",
			expected:  "LMP55.mp3",
		},
		// Standalone mode - with artist
		{
			name:      "standalone with artist",
			mode:      StandaloneMode,
			num:       "1",
			artist:    "My Show",
			cliArtist: "My Show",
			expected:  "my-show-1.mp3",
		},
		{
			name:      "standalone with artist and special chars",
			mode:      StandaloneMode,
			num:       "42",
			artist:    "The Daily Show (Late Night)",
			cliArtist: "The Daily Show (Late Night)",
			expected:  "the-daily-show-late-night-42.mp3",
		},
		{
			name:      "standalone with multiple words",
			mode:      StandaloneMode,
			num:       "99",
			artist:    "This Is A Very Long Podcast Name",
			cliArtist: "This Is A Very Long Podcast Name",
			expected:  "this-is-a-very-long-podcast-name-99.mp3",
		},
		// Standalone mode - without artist (fallback to episode)
		{
			name:      "standalone without artist",
			mode:      StandaloneMode,
			num:       "1",
			artist:    "",
			cliArtist: "",
			expected:  "episode-1.mp3",
		},
		{
			name:      "standalone without artist large number",
			mode:      StandaloneMode,
			num:       "42",
			artist:    "",
			cliArtist: "",
			expected:  "episode-42.mp3",
		},
		// Edge cases with numbers
		{
			name:      "episode number with leading zeros",
			mode:      StandaloneMode,
			num:       "007",
			artist:    "James Bond",
			cliArtist: "James Bond",
			expected:  "james-bond-007.mp3",
		},
		{
			name:      "numeric artist name",
			mode:      StandaloneMode,
			num:       "5",
			artist:    "99 Luftballons",
			cliArtist: "99 Luftballons",
			expected:  "99-luftballons-5.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global CLI.Artist to simulate the actual usage
			originalArtist := CLI.Artist
			CLI.Artist = tt.cliArtist
			defer func() { CLI.Artist = originalArtist }()

			result := generateFilename(tt.mode, tt.num, tt.artist)
			if result != tt.expected {
				t.Errorf("generateFilename(%v, %q, %q) = %q; want %q",
					tt.mode, tt.num, tt.artist, result, tt.expected)
			}
		})
	}
}

// TestResolveOutputPath tests output path resolution with directories and files
func TestResolveOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		mode       WorkflowMode
		num        string
		artist     string
		cliArtist  string
		wantErr    bool
		wantPath   string // Substring check for path validation
	}{
		// Empty output path - use current directory with generated filename
		{
			name:       "empty path uses generated filename",
			outputPath: "",
			mode:       HugoMode,
			num:        "67",
			artist:     "",
			cliArtist:  "",
			wantErr:    false,
			wantPath:   "LMP67.mp3",
		},
		{
			name:       "empty path standalone mode",
			outputPath: "",
			mode:       StandaloneMode,
			num:        "42",
			artist:     "Test Show",
			cliArtist:  "Test Show",
			wantErr:    false,
			wantPath:   "test-show-42.mp3",
		},
		// Existing directory - generate filename within it
		{
			name:       "existing directory",
			outputPath: "", // Will be set to temp dir in test
			mode:       StandaloneMode,
			num:        "1",
			artist:     "Show",
			cliArtist:  "Show",
			wantErr:    false,
			wantPath:   "show-1.mp3",
		},
		// Explicit file path - use as-is
		{
			name:       "explicit filename in current dir",
			outputPath: "custom-output.mp3",
			mode:       StandaloneMode,
			num:        "1",
			artist:     "ignored",
			cliArtist:  "ignored",
			wantErr:    false,
			wantPath:   "custom-output.mp3",
		},
		// File path in existing directory
		{
			name:       "file path in existing temp directory",
			outputPath: "", // Will be set in test
			mode:       HugoMode,
			num:        "99",
			artist:     "",
			cliArtist:  "",
			wantErr:    false,
			wantPath:   "LMP99.mp3",
		},
		// Error cases: non-existent directory
		{
			name:       "trailing slash non-existent directory",
			outputPath: "/nonexistent/dir/",
			mode:       StandaloneMode,
			num:        "1",
			artist:     "test",
			cliArtist:  "test",
			wantErr:    true,
			wantPath:   "",
		},
		{
			name:       "file in non-existent directory",
			outputPath: "/nonexistent/deeply/nested/path/file.mp3",
			mode:       StandaloneMode,
			num:        "1",
			artist:     "test",
			cliArtist:  "test",
			wantErr:    true,
			wantPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalOutputPath := CLI.OutputPath
			originalArtist := CLI.Artist
			defer func() {
				CLI.OutputPath = originalOutputPath
				CLI.Artist = originalArtist
			}()

			// Handle dynamic temp directory paths
			testOutputPath := tt.outputPath
			if tt.name == "existing directory" || tt.name == "file path in existing temp directory" {
				tmpDir := t.TempDir()
				testOutputPath = tmpDir
				if tt.name == "existing directory" {
					tt.wantPath = filepath.Join(testOutputPath, tt.wantPath)
				}
			}

			CLI.OutputPath = testOutputPath
			CLI.Artist = tt.cliArtist

			result, err := resolveOutputPath(tt.mode, tt.num, tt.artist)

			if tt.wantErr {
				if err == nil {
					t.Errorf("resolveOutputPath() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("resolveOutputPath() unexpected error: %v", err)
				return
			}

			// Check if result contains expected path component
			if tt.wantPath != "" && !isPathMatch(result, tt.wantPath) {
				t.Errorf("resolveOutputPath() = %q; want path containing %q", result, tt.wantPath)
			}
		})
	}
}

// TestResolveOutputPath_FileOverwrite tests file overwrite scenario
func TestResolveOutputPath_FileOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.mp3")

	// Create an existing file
	if err := os.WriteFile(existingFile, []byte("dummy"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Save and restore global state
	originalOutputPath := CLI.OutputPath
	defer func() { CLI.OutputPath = originalOutputPath }()

	// Resolve the same path again
	CLI.OutputPath = existingFile
	result, err := resolveOutputPath(HugoMode, "1", "")

	if err != nil {
		t.Errorf("resolveOutputPath() with existing file: got unexpected error: %v", err)
	}

	if result != existingFile {
		t.Errorf("resolveOutputPath() = %q; want %q", result, existingFile)
	}
}

// TestResolveOutputPath_GeneratedFilenameInTempDir tests generated filename placed in temp directory
func TestResolveOutputPath_GeneratedFilenameInTempDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore global state
	originalOutputPath := CLI.OutputPath
	originalArtist := CLI.Artist
	defer func() {
		CLI.OutputPath = originalOutputPath
		CLI.Artist = originalArtist
	}()

	CLI.OutputPath = tmpDir
	CLI.Artist = "Test Show"

	result, err := resolveOutputPath(StandaloneMode, "42", "Test Show")

	if err != nil {
		t.Errorf("resolveOutputPath() unexpected error: %v", err)
	}

	// Verify result is in the temp directory and has correct filename
	if !filepath.HasPrefix(result, tmpDir) {
		t.Errorf("resolveOutputPath() = %q; not in temp directory %q", result, tmpDir)
	}

	if !isPathMatch(result, "test-show-42.mp3") {
		t.Errorf("resolveOutputPath() = %q; want path containing 'test-show-42.mp3'", result)
	}
}

// isPathMatch checks if a path contains the expected component
// Handles both absolute and relative path matching
func isPathMatch(fullPath, expected string) bool {
	// Check if expected is at the end of the path (filename)
	if filepath.Base(fullPath) == expected {
		return true
	}
	// Check if expected is part of the path
	return strings.Contains(fullPath, expected)
}

// TestDetectMode tests the CLI mode detection logic for Hugo vs Standalone workflows
func TestDetectMode(t *testing.T) {
	tests := []struct {
		name      string
		audioFile string
		episodeMD string
		expected  WorkflowMode
	}{
		// Empty audio file - no arguments provided
		{
			name:      "empty audio file",
			audioFile: "",
			episodeMD: "",
			expected:  HugoMode, // Return value doesn't matter, exit will handle it
		},

		// Hugo mode: second argument is .md file
		{
			name:      "hugo mode with lowercase .md",
			audioFile: "podcast.flac",
			episodeMD: "episode.md",
			expected:  HugoMode,
		},
		{
			name:      "hugo mode with uppercase .MD",
			audioFile: "podcast.flac",
			episodeMD: "episode.MD",
			expected:  HugoMode,
		},
		{
			name:      "hugo mode with mixed case .Md",
			audioFile: "podcast.flac",
			episodeMD: "episode.Md",
			expected:  HugoMode,
		},
		{
			name:      "hugo mode with path containing .md",
			audioFile: "podcast.flac",
			episodeMD: "content/episodes/67.md",
			expected:  HugoMode,
		},
		{
			name:      "hugo mode with nested path and uppercase .MD",
			audioFile: "audio.wav",
			episodeMD: "posts/episode/post.MD",
			expected:  HugoMode,
		},
		{
			name:      "hugo mode with only filename .md",
			audioFile: "LMP67.flac",
			episodeMD: "67.md",
			expected:  HugoMode,
		},

		// Standalone mode: second argument is NOT a .md file
		{
			name:      "standalone mode with .txt file",
			audioFile: "podcast.flac",
			episodeMD: "readme.txt",
			expected:  StandaloneMode,
		},
		{
			name:      "standalone mode with .md in middle of filename",
			audioFile: "podcast.flac",
			episodeMD: "markdown_file.mp3",
			expected:  StandaloneMode,
		},
		{
			name:      "standalone mode with .md not at end",
			audioFile: "podcast.flac",
			episodeMD: "file.md.txt",
			expected:  StandaloneMode,
		},
		{
			name:      "standalone mode with empty episodeMD string",
			audioFile: "podcast.flac",
			episodeMD: "",
			expected:  StandaloneMode,
		},
		{
			name:      "standalone mode with audio file only",
			audioFile: "LMP67.flac",
			episodeMD: "",
			expected:  StandaloneMode,
		},
		{
			name:      "standalone mode with non-md extension",
			audioFile: "audio.wav",
			episodeMD: "episode.yaml",
			expected:  StandaloneMode,
		},

		// Edge cases
		{
			name:      "just .md (no filename before extension)",
			audioFile: "podcast.flac",
			episodeMD: ".md",
			expected:  HugoMode,
		},
		{
			name:      "filename with multiple dots ending in .md",
			audioFile: "podcast.flac",
			episodeMD: "my.episode.v2.md",
			expected:  HugoMode,
		},
		{
			name:      "filename with multiple dots not ending in .md",
			audioFile: "podcast.flac",
			episodeMD: "my.episode.v2.md.backup",
			expected:  StandaloneMode,
		},
		{
			name:      "markdown file with spaces",
			audioFile: "podcast.flac",
			episodeMD: "my episode file.md",
			expected:  HugoMode,
		},
		{
			name:      "episode md with special characters",
			audioFile: "podcast.flac",
			episodeMD: "episode-67_final.md",
			expected:  HugoMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalAudioFile := CLI.AudioFile
			originalEpisodeMD := CLI.EpisodeMD
			defer func() {
				CLI.AudioFile = originalAudioFile
				CLI.EpisodeMD = originalEpisodeMD
			}()

			// Set test values
			CLI.AudioFile = tt.audioFile
			CLI.EpisodeMD = tt.episodeMD

			// Call detectMode
			result := detectMode()

			// Verify result
			if result != tt.expected {
				t.Errorf("detectMode() = %v; want %v (AudioFile=%q, EpisodeMD=%q)",
					result, tt.expected, tt.audioFile, tt.episodeMD)
			}
		})
	}
}

// TestDetectMode_Integration tests detectMode in realistic scenarios
func TestDetectMode_Integration(t *testing.T) {
	tests := []struct {
		name        string
		audioFile   string
		episodeMD   string
		expected    WorkflowMode
		description string
	}{
		{
			name:        "real hugo workflow",
			audioFile:   "LMP67.flac",
			episodeMD:   "content/episodes/67.md",
			expected:    HugoMode,
			description: "User runs: jivedrop LMP67.flac content/episodes/67.md",
		},
		{
			name:        "real standalone workflow",
			audioFile:   "podcast.wav",
			episodeMD:   "",
			expected:    StandaloneMode,
			description: "User runs: jivedrop podcast.wav --title 'Ep 1' --num 1 --cover art.png",
		},
		{
			name:        "common mistake: user passes non-md file in hugo mode",
			audioFile:   "episode.flac",
			episodeMD:   "episode.txt",
			expected:    StandaloneMode,
			description: "User runs: jivedrop episode.flac episode.txt (should be .md not .txt)",
		},
		{
			name:        "edge: .md file with uppercase extension",
			audioFile:   "LMP99.flac",
			episodeMD:   "99.MD",
			expected:    HugoMode,
			description: "Handles .MD (uppercase) correctly for cross-platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalAudioFile := CLI.AudioFile
			originalEpisodeMD := CLI.EpisodeMD
			defer func() {
				CLI.AudioFile = originalAudioFile
				CLI.EpisodeMD = originalEpisodeMD
			}()

			// Set test values
			CLI.AudioFile = tt.audioFile
			CLI.EpisodeMD = tt.episodeMD

			// Call detectMode
			result := detectMode()

			// Verify result
			if result != tt.expected {
				t.Errorf("detectMode() = %v; want %v\n  Description: %s\n  AudioFile=%q, EpisodeMD=%q",
					result, tt.expected, tt.description, tt.audioFile, tt.episodeMD)
			}
		})
	}
}

// BenchmarkSanitiseForFilename benchmarks the sanitisation function
func BenchmarkSanitiseForFilename(b *testing.B) {
	testStrings := []string{
		"Linux Matters",
		"AC/DC",
		"The (Real) Show",
		"Podcast!!!???&&&",
		"Very Long Podcast Name With Many Words",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			sanitiseForFilename(s)
		}
	}
}

// BenchmarkGenerateFilename benchmarks the filename generation
func BenchmarkGenerateFilename(b *testing.B) {
	CLI.Artist = "Linux Matters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateFilename(HugoMode, "67", "")
		generateFilename(StandaloneMode, "42", "My Podcast")
		generateFilename(StandaloneMode, "1", "")
	}
}

// TestValidateHugoMode tests Hugo mode validation of episode markdown arguments
func TestValidateHugoMode(t *testing.T) {
	tests := []struct {
		name      string
		episodeMD string
		wantErr   bool
		errMatch  string // Substring to match in error message
	}{
		// Valid cases
		{
			name:      "valid markdown file lowercase .md",
			episodeMD: "episode.md",
			wantErr:   false,
		},
		{
			name:      "valid markdown file uppercase .MD",
			episodeMD: "episode.MD",
			wantErr:   false,
		},
		{
			name:      "valid markdown file mixed case .Md",
			episodeMD: "episode.Md",
			wantErr:   false,
		},
		{
			name:      "valid markdown with nested path",
			episodeMD: "content/episodes/67.md",
			wantErr:   false,
		},
		{
			name:      "valid markdown deeply nested",
			episodeMD: "posts/blog/2025/11/article.md",
			wantErr:   false,
		},
		{
			name:      "valid markdown with spaces in filename",
			episodeMD: "my episode file.md",
			wantErr:   false,
		},
		{
			name:      "valid markdown with special characters",
			episodeMD: "episode-67_final.md",
			wantErr:   false,
		},
		{
			name:      "valid markdown with multiple dots",
			episodeMD: "my.episode.v2.md",
			wantErr:   false,
		},

		// Invalid cases: missing EpisodeMD
		{
			name:      "empty episode markdown string",
			episodeMD: "",
			wantErr:   true,
			errMatch:  "requires episode markdown file",
		},

		// Invalid cases: wrong file extensions
		{
			name:      "wrong extension .txt",
			episodeMD: "episode.txt",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      "wrong extension .yaml",
			episodeMD: "episode.yaml",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      "wrong extension .json",
			episodeMD: "episode.json",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      "wrong extension .mp3",
			episodeMD: "episode.mp3",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},

		// Invalid cases: .md not at end
		{
			name:      ".md in middle of filename",
			episodeMD: "markdown_file.mp3",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      ".md with backup suffix",
			episodeMD: "episode.md.backup",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      ".md with bak extension",
			episodeMD: "episode.md.bak",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      ".md with old extension",
			episodeMD: "episode.md.old",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},

		// Edge cases
		{
			name:      "just .md filename",
			episodeMD: ".md",
			wantErr:   false,
		},
		{
			name:      "no extension",
			episodeMD: "episode",
			wantErr:   true,
			errMatch:  "must have .md extension",
		},
		{
			name:      "uppercase .MD only",
			episodeMD: ".MD",
			wantErr:   false,
		},
		{
			name:      "path with uppercase .MD",
			episodeMD: "content/episodes/POST.MD",
			wantErr:   false,
		},
		{
			name:      "relative path with ./",
			episodeMD: "./episode.md",
			wantErr:   false,
		},
		{
			name:      "relative path with ../",
			episodeMD: "../episodes/episode.md",
			wantErr:   false,
		},
		{
			name:      "absolute path (unix)",
			episodeMD: "/home/user/episodes/67.md",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalEpisodeMD := CLI.EpisodeMD
			defer func() { CLI.EpisodeMD = originalEpisodeMD }()

			// Set test value
			CLI.EpisodeMD = tt.episodeMD

			// Call validateHugoMode
			err := validateHugoMode()

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateHugoMode() expected error, got nil (EpisodeMD=%q)", tt.episodeMD)
					return
				}
				if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("validateHugoMode() error %q does not contain %q", err.Error(), tt.errMatch)
				}
				return
			}

			// Success case
			if err != nil {
				t.Errorf("validateHugoMode() unexpected error: %v (EpisodeMD=%q)", err, tt.episodeMD)
			}
		})
	}
}

// TestValidateHugoMode_Integration tests validateHugoMode with realistic scenarios
func TestValidateHugoMode_Integration(t *testing.T) {
	tests := []struct {
		name        string
		episodeMD   string
		wantErr     bool
		description string
	}{
		{
			name:        "real hugo workflow file",
			episodeMD:   "content/episodes/67.md",
			wantErr:     false,
			description: "Typical Linux Matters episode markdown path",
		},
		{
			name:        "common user mistake: .txt instead of .md",
			episodeMD:   "episode.txt",
			wantErr:     true,
			description: "User accidentally passes wrong file type",
		},
		{
			name:        "common user mistake: no extension",
			episodeMD:   "episode",
			wantErr:     true,
			description: "User passes file without extension",
		},
		{
			name:        "common user mistake: .md.bak backup file",
			episodeMD:   "episode.md.bak",
			wantErr:     true,
			description: "User passes backup file instead of original",
		},
		{
			name:        "windows path with backslashes",
			episodeMD:   "content\\episodes\\67.md",
			wantErr:     false,
			description: "Cross-platform support for Windows paths",
		},
		{
			name:        "uppercase .MD extension",
			episodeMD:   "EPISODE.MD",
			wantErr:     false,
			description: "Case-insensitive extension matching",
		},
		{
			name:        "missing argument",
			episodeMD:   "",
			wantErr:     true,
			description: "User forgot to provide episode markdown file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalEpisodeMD := CLI.EpisodeMD
			defer func() { CLI.EpisodeMD = originalEpisodeMD }()

			// Set test value
			CLI.EpisodeMD = tt.episodeMD

			// Call validateHugoMode
			err := validateHugoMode()

			// Check error expectations
			if tt.wantErr && err == nil {
				t.Errorf("validateHugoMode() expected error but got nil\n  Description: %s\n  EpisodeMD=%q",
					tt.description, tt.episodeMD)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateHugoMode() unexpected error: %v\n  Description: %s\n  EpisodeMD=%q",
					err, tt.description, tt.episodeMD)
			}
		})
	}
}

// TestValidateStandaloneMode tests standalone mode validation of required flags
func TestValidateStandaloneMode(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		num      string
		cover    string
		wantErr  bool
		errMatch string // Substring to match in error message
	}{
		// Valid cases: all flags present
		{
			name:    "all flags present",
			title:   "Episode Title",
			num:     "1",
			cover:   "cover.png",
			wantErr: false,
		},
		{
			name:    "all flags with special characters",
			title:   "Episode: The Quest (Part 1)",
			num:     "42",
			cover:   "artwork/cover-2025.png",
			wantErr: false,
		},
		{
			name:    "all flags with long values",
			title:   "This is a very long episode title with lots of words and details",
			num:     "999",
			cover:   "/absolute/path/to/very/deep/directory/structure/cover.png",
			wantErr: false,
		},
		{
			name:    "all flags with minimal values",
			title:   "A",
			num:     "0",
			cover:   "c.png",
			wantErr: false,
		},
		{
			name:    "all flags with unicode in title",
			title:   "Podcast — Episode 1",
			num:     "1",
			cover:   "cover.png",
			wantErr: false,
		},

		// Invalid cases: missing title
		{
			name:     "missing title",
			title:    "",
			num:      "1",
			cover:    "cover.png",
			wantErr:  true,
			errMatch: "requires --title flag",
		},
		{
			name:     "missing title with num and cover",
			title:    "",
			num:      "42",
			cover:    "/path/to/cover.png",
			wantErr:  true,
			errMatch: "requires --title flag",
		},

		// Invalid cases: missing num
		{
			name:     "missing num",
			title:    "Episode Title",
			num:      "",
			cover:    "cover.png",
			wantErr:  true,
			errMatch: "requires --num flag",
		},
		{
			name:     "missing num with title and cover",
			title:    "My Show",
			num:      "",
			cover:    "artwork.png",
			wantErr:  true,
			errMatch: "requires --num flag",
		},

		// Invalid cases: missing cover
		{
			name:     "missing cover",
			title:    "Episode Title",
			num:      "1",
			cover:    "",
			wantErr:  true,
			errMatch: "requires --cover flag",
		},
		{
			name:     "missing cover with title and num",
			title:    "My Podcast",
			num:      "100",
			cover:    "",
			wantErr:  true,
			errMatch: "requires --cover flag",
		},

		// Invalid cases: multiple flags missing
		{
			name:     "missing title and num",
			title:    "",
			num:      "",
			cover:    "cover.png",
			wantErr:  true,
			errMatch: "requires --title flag", // First validation fails
		},
		{
			name:     "missing title and cover",
			title:    "",
			num:      "1",
			cover:    "",
			wantErr:  true,
			errMatch: "requires --title flag", // First validation fails
		},
		{
			name:     "missing num and cover",
			title:    "Episode",
			num:      "",
			cover:    "",
			wantErr:  true,
			errMatch: "requires --num flag", // First validation fails
		},
		{
			name:     "all flags empty",
			title:    "",
			num:      "",
			cover:    "",
			wantErr:  true,
			errMatch: "requires --title flag", // First validation fails
		},

		// Edge cases: whitespace-only values (still invalid if they parse as empty)
		{
			name:    "whitespace-only title (not trimmed by validator)",
			title:   "   ",
			num:     "1",
			cover:   "cover.png",
			wantErr: false, // Not empty string, so passes validation
		},
		{
			name:    "whitespace-only num (not trimmed by validator)",
			title:   "Episode",
			num:     "   ",
			cover:   "cover.png",
			wantErr: false, // Not empty string, so passes validation
		},

		// Edge cases: numeric-like values
		{
			name:    "num as string with leading zeros",
			title:   "Episode",
			num:     "007",
			cover:   "cover.png",
			wantErr: false,
		},
		{
			name:    "num as negative (still valid as string)",
			title:   "Episode",
			num:     "-1",
			cover:   "cover.png",
			wantErr: false,
		},
		{
			name:    "num as decimal (valid as string)",
			title:   "Episode",
			num:     "1.5",
			cover:   "cover.png",
			wantErr: false,
		},

		// Cover art path variations
		{
			name:    "cover as url",
			title:   "Episode",
			num:     "1",
			cover:   "https://example.com/cover.png",
			wantErr: false,
		},
		{
			name:    "cover as relative path",
			title:   "Episode",
			num:     "1",
			cover:   "./images/cover.png",
			wantErr: false,
		},
		{
			name:    "cover as absolute path",
			title:   "Episode",
			num:     "1",
			cover:   "/absolute/path/cover.png",
			wantErr: false,
		},
		{
			name:    "cover with spaces in path",
			title:   "Episode",
			num:     "1",
			cover:   "my cover art/final version.png",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalTitle := CLI.Title
			originalNum := CLI.Num
			originalCover := CLI.Cover
			defer func() {
				CLI.Title = originalTitle
				CLI.Num = originalNum
				CLI.Cover = originalCover
			}()

			// Set test values
			CLI.Title = tt.title
			CLI.Num = tt.num
			CLI.Cover = tt.cover

			// Call validateStandaloneMode
			err := validateStandaloneMode()

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateStandaloneMode() expected error, got nil\n  Title=%q, Num=%q, Cover=%q",
						tt.title, tt.num, tt.cover)
					return
				}
				if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("validateStandaloneMode() error %q does not contain %q", err.Error(), tt.errMatch)
				}
				return
			}

			// Success case
			if err != nil {
				t.Errorf("validateStandaloneMode() unexpected error: %v\n  Title=%q, Num=%q, Cover=%q",
					err, tt.title, tt.num, tt.cover)
			}
		})
	}
}

// TestValidateStandaloneMode_Integration tests validateStandaloneMode with realistic scenarios
func TestValidateStandaloneMode_Integration(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		num         string
		cover       string
		wantErr     bool
		description string
	}{
		{
			name:        "valid standalone workflow",
			title:       "Terminal Full of Sparkles",
			num:         "66",
			cover:       "artwork.png",
			wantErr:     false,
			description: "User runs: jivedrop audio.flac --title 'Terminal Full of Sparkles' --num 66 --cover artwork.png",
		},
		{
			name:        "common mistake: forgot --title flag",
			title:       "",
			num:         "42",
			cover:       "cover.png",
			wantErr:     true,
			description: "User forgets required --title flag",
		},
		{
			name:        "common mistake: forgot --num flag",
			title:       "My Episode",
			num:         "",
			cover:       "cover.png",
			wantErr:     true,
			description: "User forgets required --num flag (episode number)",
		},
		{
			name:        "common mistake: forgot --cover flag",
			title:       "My Episode",
			num:         "1",
			cover:       "",
			wantErr:     true,
			description: "User forgets required --cover flag",
		},
		{
			name:        "forgot all flags",
			title:       "",
			num:         "",
			cover:       "",
			wantErr:     true,
			description: "User forgets all required flags",
		},
		{
			name:        "minimal valid values",
			title:       "Ep",
			num:         "1",
			cover:       "art.png",
			wantErr:     false,
			description: "Minimal valid values for standalone mode",
		},
		{
			name:        "episode with complex path",
			title:       "The Daily Show",
			num:         "99",
			cover:       "/home/user/podcasts/2025/daily-show-99.png",
			wantErr:     false,
			description: "Real-world example with full path to cover",
		},
		{
			name:        "missing only cover field",
			title:       "Podcast Name",
			num:         "50",
			cover:       "",
			wantErr:     true,
			description: "Most common mistake: user specifies --title and --num but forgets --cover",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			originalTitle := CLI.Title
			originalNum := CLI.Num
			originalCover := CLI.Cover
			defer func() {
				CLI.Title = originalTitle
				CLI.Num = originalNum
				CLI.Cover = originalCover
			}()

			// Set test values
			CLI.Title = tt.title
			CLI.Num = tt.num
			CLI.Cover = tt.cover

			// Call validateStandaloneMode
			err := validateStandaloneMode()

			// Check error expectations
			if tt.wantErr && err == nil {
				t.Errorf("validateStandaloneMode() expected error but got nil\n  Description: %s\n  Title=%q, Num=%q, Cover=%q",
					tt.description, tt.title, tt.num, tt.cover)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateStandaloneMode() unexpected error: %v\n  Description: %s\n  Title=%q, Num=%q, Cover=%q",
					err, tt.description, tt.title, tt.num, tt.cover)
			}
		})
	}
}
