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
