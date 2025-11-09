package encoder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEpisodeMetadata(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid frontmatter",
			content: `---
episode: "67"
title: "Mirrors, Motors and Makefiles"
Date: 2025-11-09T00:00:00Z
episode_image: "/img/episode/linuxmatters-3000x3000.png"
---

Episode content here.
`,
			wantErr: false,
		},
		{
			name: "missing episode field",
			content: `---
title: "Test Episode"
Date: 2025-11-09T00:00:00Z
episode_image: "/img/test.png"
---
`,
			wantErr:     true,
			errContains: "missing required field: episode",
		},
		{
			name: "missing title field",
			content: `---
episode: "67"
Date: 2025-11-09T00:00:00Z
episode_image: "/img/test.png"
---
`,
			wantErr:     true,
			errContains: "missing required field: title",
		},
		{
			name: "missing episode_image field",
			content: `---
episode: "67"
title: "Test Episode"
Date: 2025-11-09T00:00:00Z
---
`,
			wantErr:     true,
			errContains: "missing required field: episode_image",
		},
		{
			name: "invalid YAML",
			content: `---
episode: "67"
title: [invalid: yaml: syntax
Date: 2025-11-09T00:00:00Z
episode_image: "/img/test.png"
---
`,
			wantErr:     true,
			errContains: "failed to parse frontmatter",
		},
		{
			name: "missing closing delimiter",
			content: `---
episode: "67"
title: "Test Episode"
Date: 2025-11-09T00:00:00Z
episode_image: "/img/test.png"

No closing delimiter
`,
			wantErr:     true,
			errContains: "invalid frontmatter",
		},
		{
			name:        "no frontmatter delimiters",
			content:     "Just plain content without frontmatter\n",
			wantErr:     true,
			errContains: "invalid frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the metadata
			meta, err := ParseEpisodeMetadata(tmpFile)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errContains, err)
				}
				return
			}

			// Check success expectations
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if meta == nil {
				t.Error("Expected metadata, got nil")
			}
		})
	}
}

func TestParseEpisodeMetadata_ValidFields(t *testing.T) {
	content := `---
episode: "67"
title: "Mirrors, Motors and Makefiles"
Date: 2025-11-09T00:00:00Z
episode_image: "/img/episode/linuxmatters-3000x3000.png"
podcast_duration: "54:09"
podcast_bytes: 27357184
---

Episode content.
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	meta, err := ParseEpisodeMetadata(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all fields
	if meta.Episode != "67" {
		t.Errorf("Expected episode '67', got '%s'", meta.Episode)
	}
	if meta.Title != "Mirrors, Motors and Makefiles" {
		t.Errorf("Expected title 'Mirrors, Motors and Makefiles', got '%s'", meta.Title)
	}
	if meta.EpisodeImage != "/img/episode/linuxmatters-3000x3000.png" {
		t.Errorf("Expected episode_image '/img/episode/linuxmatters-3000x3000.png', got '%s'", meta.EpisodeImage)
	}
	if meta.PodcastDuration != "54:09" {
		t.Errorf("Expected podcast_duration '54:09', got '%s'", meta.PodcastDuration)
	}
	if meta.PodcastBytes != 27357184 {
		t.Errorf("Expected podcast_bytes 27357184, got %d", meta.PodcastBytes)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
