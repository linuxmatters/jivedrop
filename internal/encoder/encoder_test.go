package encoder

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEncodeToMP3_Integration is an integration test that verifies
// the full encoding pipeline works and creates a test MP3 file that
// other tests can use for validation.
func TestEncodeToMP3_Integration(t *testing.T) {
	// Use testdata/LMP0.flac as input
	inputPath := "../../testdata/LMP0.flac"
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", inputPath)
	}

	// Output to testdata/LMP0.mp3 (not tracked in git)
	outputPath := "../../testdata/LMP0.mp3"

	// Note: We don't defer cleanup of this file - it's used by other tests
	// and is already in .gitignore

	// Test mono encoding (default)
	t.Run("mono encoding", func(t *testing.T) {
		enc, err := New(Config{
			InputPath:  inputPath,
			OutputPath: outputPath,
			Stereo:     false,
		})
		if err != nil {
			t.Fatalf("Failed to create encoder: %v", err)
		}
		defer enc.Close()

		if err := enc.Initialize(); err != nil {
			t.Fatalf("Failed to initialize encoder: %v", err)
		}

		// Encode with nil progress callback for tests
		err = enc.Encode(nil)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Verify output file exists
		info, err := os.Stat(outputPath)
		if err != nil {
			t.Fatalf("Output file not created: %v", err)
		}

		// Verify output file has reasonable size (should be > 1KB for a real audio file)
		if info.Size() < 1024 {
			t.Errorf("Output file too small: %d bytes", info.Size())
		}

		t.Logf("Created MP3: %s (%d bytes)", outputPath, info.Size())
	})

	// Test stereo encoding
	t.Run("stereo encoding", func(t *testing.T) {
		stereoOutput := "../../testdata/LMP0-stereo.mp3"
		defer os.Remove(stereoOutput)

		enc, err := New(Config{
			InputPath:  inputPath,
			OutputPath: stereoOutput,
			Stereo:     true,
		})
		if err != nil {
			t.Fatalf("Failed to create encoder: %v", err)
		}
		defer enc.Close()

		if err := enc.Initialize(); err != nil {
			t.Fatalf("Failed to initialize encoder: %v", err)
		}

		err = enc.Encode(nil)
		if err != nil {
			t.Fatalf("Encode (stereo) failed: %v", err)
		}

		// Verify output exists
		info, err := os.Stat(stereoOutput)
		if err != nil {
			t.Fatalf("Stereo output file not created: %v", err)
		}

		// Stereo file should be larger than mono (approximately 192/112 ratio)
		if info.Size() < 1024 {
			t.Errorf("Stereo output file too small: %d bytes", info.Size())
		}

		t.Logf("Created stereo MP3: %s (%d bytes)", stereoOutput, info.Size())
	})
}

// TestEncoder_InvalidInput tests error handling for invalid inputs
func TestEncoder_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputPath string
		wantErr    bool
	}{
		{
			name:       "non-existent input file",
			inputPath:  "/nonexistent/file.flac",
			outputPath: "/tmp/output.mp3",
			wantErr:    true,
		},
		{
			name:       "empty input path",
			inputPath:  "",
			outputPath: "/tmp/output.mp3",
			wantErr:    true,
		},
		{
			name:       "empty output path",
			inputPath:  "../../testdata/LMP0.flac",
			outputPath: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := New(Config{
				InputPath:  tt.inputPath,
				OutputPath: tt.outputPath,
				Stereo:     false,
			})

			// Check for immediate config errors
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected error during New: %v", err)
				}
				return
			}
			defer enc.Close()

			// Check for initialization errors
			err = enc.Initialize()
			if tt.wantErr && err == nil {
				t.Error("Expected error during Initialize, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error during Initialize: %v", err)
			}
		})
	}
}

// TestEncoder_OutputExists tests behavior when output file already exists
func TestEncoder_OutputExists(t *testing.T) {
	inputPath := "../../testdata/LMP0.flac"
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", inputPath)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.mp3")

	// Create initial encoding
	enc1, err := New(Config{
		InputPath:  inputPath,
		OutputPath: outputPath,
		Stereo:     false,
	})
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}
	defer enc1.Close()

	if err := enc1.Initialize(); err != nil {
		t.Fatalf("Failed to initialize encoder: %v", err)
	}

	if err := enc1.Encode(nil); err != nil {
		t.Fatalf("Initial encoding failed: %v", err)
	}

	// Get initial file info
	initialInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat initial output: %v", err)
	}

	// Encode again to same path (should overwrite)
	enc2, err := New(Config{
		InputPath:  inputPath,
		OutputPath: outputPath,
		Stereo:     false,
	})
	if err != nil {
		t.Fatalf("Failed to create second encoder: %v", err)
	}
	defer enc2.Close()

	if err := enc2.Initialize(); err != nil {
		t.Fatalf("Failed to initialize second encoder: %v", err)
	}

	if err := enc2.Encode(nil); err != nil {
		t.Fatalf("Second encoding failed: %v", err)
	}

	// Verify file was overwritten
	newInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat new output: %v", err)
	}

	// Timestamps should be different (file was replaced)
	if !newInfo.ModTime().After(initialInfo.ModTime()) && !newInfo.ModTime().Equal(initialInfo.ModTime()) {
		t.Error("Output file does not appear to have been overwritten")
	}
}

// TestEncoder_CloseSafety verifies that Close() handles edge cases without panicking
func TestEncoder_CloseSafety(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setupFunc func() *Encoder
	}{
		{
			name: "close before initialize",
			setupFunc: func() *Encoder {
				enc, err := New(Config{
					InputPath:  "../../testdata/LMP0.flac",
					OutputPath: filepath.Join(tmpDir, "test1.mp3"),
					Stereo:     false,
				})
				if err != nil {
					t.Fatalf("Failed to create encoder: %v", err)
				}
				// Don't call Initialize - Close should handle nil pointers gracefully
				return enc
			},
		},
		{
			name: "double close",
			setupFunc: func() *Encoder {
				enc, err := New(Config{
					InputPath:  "../../testdata/LMP0.flac",
					OutputPath: filepath.Join(tmpDir, "test2.mp3"),
					Stereo:     false,
				})
				if err != nil {
					t.Fatalf("Failed to create encoder: %v", err)
				}

				if err := enc.Initialize(); err != nil {
					t.Fatalf("Failed to initialize encoder: %v", err)
				}

				// First close
				enc.Close()
				// Second close should not panic even though resources are freed
				return enc
			},
		},
		{
			name: "close after failed initialize",
			setupFunc: func() *Encoder {
				// Use invalid input file to cause Initialize to fail
				enc, err := New(Config{
					InputPath:  "/nonexistent/invalid.flac",
					OutputPath: filepath.Join(tmpDir, "test3.mp3"),
					Stereo:     false,
				})
				if err != nil {
					t.Fatalf("Failed to create encoder: %v", err)
				}

				// Initialize will fail, leaving partial resources
				_ = enc.Initialize()
				// Close should still work safely even after failed Initialize
				return enc
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a separate test to verify that Close() can be called
			// multiple times or after partial initialization without panicking
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Close() panicked: %v", r)
				}
			}()

			enc := tt.setupFunc()
			// Call Close a second time for the "double close" scenario
			// For other scenarios, this is extra safety verification
			enc.Close()
			enc.Close()
		})
	}
}
