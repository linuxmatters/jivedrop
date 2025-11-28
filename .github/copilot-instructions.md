# Jivedrop AI Agent Instructions

Jivedrop is a Go CLI tool that encodes podcast audio (WAV/FLAC) to MP3 with embedded ID3v2 tags and cover art for RSS distribution.

## Architecture

```
cmd/jivedrop/main.go     # CLI entry, mode detection (Hugo vs Standalone), argument validation
internal/
  encoder/               # FFmpeg-based MP3 encoding via ffmpeg-statigo
    encoder.go           # Core encode pipeline: decode → filter → encode
    metadata.go          # Hugo frontmatter parsing (YAML between --- delimiters)
    stats.go             # Duration/filesize extraction from encoded MP3
  id3/                   # ID3v2.4 tag writing via bogem/id3v2
    writer.go            # Tag frames: TIT2, TALB, TPE1, TDRC, COMM, APIC
    artwork.go           # Cover art scaling (1400-3000px range for Apple Podcasts)
  ui/                    # Bubbletea TUI for encoding progress
    encode.go            # Progress model with realtime speed calculation
  cli/                   # Lipgloss-styled output
    help.go              # Custom Kong help printer
    styles.go            # Colour palette (matches Jivefire sibling project)
third_party/ffmpeg-statigo/  # Git submodule: FFmpeg 8.0 static bindings
```

This project uses `ffmpeg-statigo` for FFmpeg 8.0 static bindings, included as a git submodule in `third_party/ffmpeg-statigo`. Key locations within the submodule:
- `*.gen.go` files (e.g., `functions.gen.go`, `structs.gen.go`) - auto-generated Go bindings, do not edit
- `include/` - FFmpeg C headers used for CGO compilation

## Development Commands

**Always use `just` commands** - never run `go build` or `go test` directly:

```bash
just setup      # REQUIRED first: init submodule + download FFmpeg libs
just build      # Build binary with version from git tags
just test       # Run all tests
just mp3        # Integration test: encode testdata/LMP67.flac
just clean      # Remove build artifacts and test outputs
```

## Key Patterns

### Dual-Mode CLI
- **Hugo mode**: `jivedrop audio.flac episode.md` - reads metadata from Hugo frontmatter
- **Standalone mode**: `jivedrop audio.flac --title X --num N --cover Y` - explicit flags
- Mode detection: second arg ending `.md` triggers Hugo mode

### Hugo Frontmatter
- Required fields in episode markdown: `episode`, `title`, `episode_image`
- After encoding, Jivedrop calculates `podcast_duration` and `podcast_bytes`
- Prompts user to update frontmatter if values differ or are missing

### Encoding Settings
- **Mono (default)**: CBR 112kbps, 44.1kHz, LAME quality 3, 20.5kHz lowpass
- **Stereo (`--stereo`)**: CBR 192kbps, 44.1kHz, LAME quality 3, 20.5kHz lowpass

### FFmpeg Integration
- Uses `ffmpeg-statigo` submodule for static FFmpeg bindings (no system FFmpeg needed)
- Filter graph: resample → channel downmix → LAME encode
- All FFmpeg types prefixed with `ffmpeg.AV*`

### Error Handling
- Use `cli.PrintError()` and `cli.PrintInfo()` for user-facing messages
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Clean up partial files on encoding failure

### Test Data
- `testdata/` contains sample FLAC, markdown, and artwork
- Tests output to `testdata/*.mp3` (cleaned by `just clean`)

## Environment
- NixOS development shell via `flake.nix`
- Fish shell for terminal commands
- CGO required (`CGO_ENABLED=1` in build)

## Conventions

- British English spelling in user-facing text and comments
- Lipgloss styles in `internal/cli/styles.go` define the colour palette
- Kong for CLI parsing with custom help printer
- Bubbletea for interactive progress UI during encoding
