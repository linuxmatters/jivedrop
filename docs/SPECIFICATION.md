# Jivedrop Specification 🪩

**MP3 encoder for podcast distribution with embedded artwork and ID3 metadata.**

Version: 0.1.0
Status: Specification
Date: 9 November 2025

---

## Overview

Jivedrop encodes podcast audio (WAV, FLAC) to optimized MP3 files for RSS feed distribution. It embeds cover artwork and ID3v2 metadata, integrates with the Linux Matters Hugo website workflow, and outputs podcast frontmatter values.

### The Problem

`mp3-encode.sh` works but has limitations:
- Bash script requires multiple external tools (LAME, kid3-cli, mediainfo)
- Dependency management fragile across systems
- No progress indication during encoding
- Error handling basic
- Not portable (platform-specific tool availability)

### The Solution

Go single binary using ffmpeg-go with:
- Embedded FFmpeg libraries (LAME encoder included)
- ID3v2 tag writing
- Cover art embedding
- Bubbletea UI with encoding progress
- Cross-platform (Linux, macOS, Windows)
- Zero external dependencies

---

## Core Design Principles

### 1. Single-Binary Distribution
Static FFmpeg libraries via ffmpeg-go (same approach as Jivefire). No LAME, kid3-cli, or mediainfo installation required.

### 2. Linux Matters Workflow Integration
- Read episode metadata from Hugo markdown files (`content/episode/67.md`)
- Locate cover art from episode frontmatter
- Output podcast frontmatter values (duration, bytes)
- Follow existing naming conventions (`LMP{num}.mp3`)

### 3. Professional UX
Bubbletea UI showing:
- Encoding progress (percentage, time elapsed/remaining)
- Audio specifications (sample rate, channels, bitrate)
- Real-time bitrate monitoring
- Output file statistics

### 4. Sensible Defaults, Full Control
Default LAME settings optimized for podcast distribution:
- **Mono (default):** CBR 112kbps, 44.1kHz, quality 3, 20.5kHz lowpass
- **Stereo (--stereo flag):** CBR 192kbps, 44.1kHz, quality 3, 20.5kHz lowpass

No unnecessary options—just `--stereo` for higher quality stereo encoding.

---

## Technical Architecture

### Encoding Pipeline

```
Input Audio (WAV/FLAC)
    ↓
Decode (ffmpeg-go)
    ↓
Resample to 44.1kHz
    ↓
Downmix to mono
    ↓
Encode to MP3 (LAME via ffmpeg-go)
    ↓
Write ID3v2 tags
    ↓
Embed cover art (APIC frame)
    ↓
Output MP3 file
```

### ID3v2 Tag Structure

```
TIT2: {episode_num}: {episode_title}
TALB: Linux Matters
TRCK: {episode_num}
TPE1: Linux Matters
TDRC: {current_year}-{current_month}  (e.g., "2025-11")
COMM: https://linuxmatters.sh/
APIC: Cover art (PNG, front cover, scaled per requirements)
```

### Metadata Extraction

Parse Hugo frontmatter from episode markdown file. Support Hugo's YAML format (delimited by `---`):

```yaml
---
episode: 67
title: "Mirrors, Motors and Makefiles"
podcast_image: "/img/episode/linuxmatters-3000x3000.png"
podcast_duration: ""  # Jivedrop fills this
podcast_bytes: 0      # Jivedrop fills this
---
```

**Required frontmatter fields:**
- `episode`: Episode number (integer)
- `title`: Episode title (string)
- `podcast_image`: Path to cover art (string, relative to website root)

**Optional frontmatter fields:**
- `podcast_duration`: Filled by Jivedrop after encoding
- `podcast_bytes`: Filled by Jivedrop after encoding

### File Statistics

Use ffmpeg-go to extract:
- Duration: HH:MM:SS format (matching mediainfo output)
- File size: bytes for frontmatter `podcast_bytes` field

---

## CLI Interface

### Basic Encoding
```bash
./jivedrop /path/to/epidode.md /path/to/LMP67.wav
```

### Custom Output Directory
```bash
./jivedrop /path/to/epidode.md /path/to/LMP67.wav /tmp
```

### Arguments

```
jivedrop [FLAGS] <episode-md> <input-file> [output-dir]

ARGUMENTS:
  episode-md        Path to episode markdown file (e.g., 67.md)
  input-file        Path to audio file (WAV, FLAC)
  output-dir        Optional output directory (default: current directory)

FLAGS:
  --stereo          Encode as stereo at 192kbps (default: mono at 112kbps)
  --cover PATH      Custom cover art path (overrides frontmatter)
  --version         Show version and exit
  --help            Show this help and exit

OUTPUT:
  Creates LMP{num}.mp3 where {num} is from frontmatter with:
  - Embedded ID3v2 tags (episode, title, cover art, date)
  - Optimized encoding (44.1kHz, CBR, quality preset 3, 20.5kHz lowpass)
  - Displays frontmatter values (podcast_duration, podcast_bytes)

ENCODING SETTINGS:
  Mono:   112kbps CBR, 44.1kHz, quality 3, 20.5kHz lowpass
  Stereo: 192kbps CBR, 44.1kHz, quality 3, 20.5kHz lowpass
```

### Example Output

```
Jivedrop 🪩 v0.1.0

Episode: 67 - Mirrors, Motors and Makefiles
Input:   /path/to/LMP67.wav (PCM 48kHz stereo)
Output:  LMP67.mp3

Encoding to MP3...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100%

Specs:   CBR 112kbps, 44.1kHz mono
Quality: -q 3, lowpass 20.5kHz
Encoded: 3249.6s (54:09) in 42.3s (76.8x realtime)

Embedding ID3v2 tags...
  Title:   67: Mirrors, Motors and Makefiles
  Album:   Linux Matters
  Track:   67
  Artist:  Linux Matters
  Date:    2025-11
  Comment: https://linuxmatters.sh/
  Cover:   1400x1400 PNG (128KB)

✓ Complete: LMP67.mp3

Podcast frontmatter:
  podcast_duration: 54:09
  podcast_bytes: 27357184
```

---

## File Structure

```
jivedrop/
├── cmd/
│   └── jivedrop/
│       └── main.go           # Entry point
├── internal/
│   ├── cli/                  # Shared with Jivefire
│   │   ├── help.go          # Styled help output
│   │   └── styles.go        # Lipgloss colour palette
│   ├── encoder/
│   │   ├── encoder.go       # ffmpeg-go MP3 encoding
│   │   └── metadata.go      # Hugo markdown parsing
│   ├── id3/
│   │   ├── writer.go        # ID3v2 tag writing
│   │   └── artwork.go       # APIC frame embedding
│   └── ui/
│       ├── encode.go        # Bubbletea encoding model
│       └── styles.go        # Progress bar, spinners
├── docs/
│   └── SPECIFICATION.md     # This file
├── testdata/
│   ├── LMP67.wav           # Test audio
│   └── test.png            # Test artwork
├── flake.nix               # Nix development shell
├── go.mod
├── go.sum
├── justfile                # Build automation
├── LICENSE                 # GPL-3.0 (ffmpeg-go requirement)
└── README.md
```

---

## Implementation Phases

### Phase 0: Foundation
**Goal:** Project structure, dependencies, basic CLI

- [x] Go module initialization
- [x] Kong CLI argument parsing
- [x] Styled help output (Lipgloss)
- [x] Version flag
- [x] Input validation (file exists, episode markdown exists)

**Success criteria:**
- ✅ `jivedrop --help` shows styled help
- ✅ `jivedrop --version` shows version
- ✅ File validation rejects missing files

---

### Phase 1: Core Encoding
**Goal:** MP3 encoding via ffmpeg-go

- [ ] ffmpeg-go integration
- [ ] Audio decoder (WAV, FLAC input)
- [ ] LAME encoder configuration
  - CBR 112kbps
  - 44.1kHz resampling
  - Mono downmix
  - Quality preset `-q 3`
  - Lowpass 20.5kHz
- [ ] Basic progress output (no UI yet)

**Success criteria:**
- Encodes WAV to MP3 with correct settings
- Matches current `lame` output quality
- File plays correctly in media players

---

### Phase 2: Metadata Integration
**Goal:** Hugo workflow integration and ID3 tags

- [ ] Hugo markdown parser
  - Extract episode title, number, date and path to `episode_image` from the frontmatter in the episode markdown.
  - Validate episode file exists
- [ ] ID3v2 tag writing
  - TIT2 (title): `{num}: {title}`
  - TALB (album): `Linux Matters`
  - TRCK (track): `{num}`
  - TPE1 (artist): `Linux Matters`
  - TDRC (date): year and month
  - COMM (comment): website URL
- [ ] Cover art embedding (APIC frame)

**Library options:**
- `github.com/bogem/id3v2` (active, pure Go, clean API)
- `github.com/dhowden/tag` (read/write, broader format support)

**Success criteria:**
- Episode title correctly extracted from markdown
- ID3 tags visible in media players
- Cover art displays in players

---

### Phase 3: Bubbletea UI
**Goal:** Professional progress indication

- [ ] Encoding progress model
  - Percentage complete
  - Time elapsed / remaining
  - Real-time bitrate
  - Audio specs display
- [ ] Tag embedding spinner
- [ ] Success/error states
- [ ] Styled output matching Jivefire aesthetic

**UI mockup:**
```
┌─────────────────────────────────────────────────┐
│ Encoding to MP3...                              │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 73%              │
│                                                 │
│ Elapsed:  32.1s | Remaining: ~12.4s             │
│ Bitrate:  112 kbps | Speed: 101.2x realtime     │
│                                                 │
│ Input:   PCM 48kHz stereo                       │
│ Output:  MP3 44.1kHz mono CBR 112kbps           │
└─────────────────────────────────────────────────┘
```

**Success criteria:**
- Smooth progress updates (no flicker)
- Accurate time estimation
- Clean UI rendering

---

### Phase 4: File Statistics
**Goal:** Output podcast frontmatter values

- [ ] Extract duration from encoded MP3
  - Format as HH:MM:SS (mediainfo compatible)
- [ ] Calculate file size in bytes
- [ ] Display frontmatter output
  - `podcast_duration: HH:MM:SS`
  - `podcast_bytes: NNNNN`

**Success criteria:**
- Duration matches mediainfo output
- Byte count matches `du -sb` output
- Frontmatter copy-pasteable into Hugo

---

### Phase 5: Production Hardening
**Goal:** Error handling, edge cases, polish

- [ ] Error handling
  - Missing cover art (error with path)
  - Invalid episode markdown (error with example)
  - Encoding failures (cleanup partial files)
  - Disk full scenarios
- [ ] Overwrite protection
  - Detect existing output file
  - Prompt for confirmation (or `--force` flag)
- [ ] CLI flag validation
  - Bitrate range (64-320 kbps reasonable)
  - Quality range (0-9)
  - Sample rate validation (8000-192000 Hz)
- [ ] Testing
  - Unit tests for metadata extraction
  - Integration tests for encoding pipeline
  - Test with various input formats

**Success criteria:**
- No panics on bad input
- Clear error messages
- Partial files cleaned up on failure
- All tests passing

---

## Dependencies

### Core
```go
github.com/csnewman/ffmpeg-go v0.6.0         // MP3 encoding
github.com/charmbracelet/bubbletea v1.3.10   // TUI framework
github.com/charmbracelet/lipgloss v1.1.0     // Styling
github.com/alecthomas/kong v1.12.1           // CLI parsing
```

### ID3 Tags
```go
github.com/bogem/id3v2 v1.2.0                // ID3v2 tag writing
```

### Metadata Parsing
```go
github.com/BurntSushi/toml v1.4.0            // Hugo frontmatter (if TOML)
gopkg.in/yaml.v3 v3.0.1                      // Hugo frontmatter (if YAML)
```

Standard library sufficient for:
- File I/O
- Regex for title extraction
- Duration formatting

---

## LAME Encoder Settings

### Default Configuration
Match `mp3-encode.sh` exactly:

```bash
-a              # Auto-select channel mode
-m m            # Force mono mode
-q 3            # Quality preset 3 (high quality)
--lowpass 20.5  # Lowpass filter at 20.5kHz
-b 112          # Constant bitrate 112kbps
--cbr           # Force constant bitrate
--resample 44.1 # Resample to 44.1kHz
--noreplaygain  # Disable ReplayGain analysis
```

### ffmpeg-go Translation

```go
// Encoder options dictionary
opts := ffmpeg.AVDictSet(nil,
    "b:a", "112k",           // Audio bitrate
    "ar", "44100",           // Audio sample rate
    "ac", "1",               // Audio channels (mono)
    "compression_level", "3", // LAME quality
    "lowpass", "20500",      // Lowpass frequency (Hz)
    nil,
)

// Configure LAME encoder
codec := ffmpeg.AVCodecFindEncoder(ffmpeg.AVCodecIdMP3)
codecContext := ffmpeg.AVCodecAllocContext3(codec)
codecContext.SetBitRate(112000)
codecContext.SetSampleRate(44100)
codecContext.SetChannels(1)
// ... open with opts dictionary
```

---

## ID3v2 Tag Details

### Frame Specifications

```
Frame: TIT2 (Title)
  Value: "{episode_num}: {episode_title}"
  Example: "67: Mirrors, Motors and Makefiles"

Frame: TALB (Album)
  Value: "Linux Matters"

Frame: TRCK (Track number)
  Value: "{episode_num}"
  Example: "67"

Frame: TPE1 (Lead artist)
  Value: "Linux Matters"

Frame: TDRC (Recording date)
  Value: "{year}-{month}"
  Example: "2025-11"

Frame: COMM (Comment)
  Language: "eng"
  Description: ""
  Value: "https://linuxmatters.sh/"

Frame: APIC (Attached picture)
  MIME: "image/png"
  Picture type: 3 (Cover front)
  Description: "Linux Matters Logo"
  Data: PNG image bytes
```

### Cover Art Requirements

- Format: PNG (existing artwork is PNG)
- Scaling logic:
  - Input > 3000x3000: Scale down to 3000x3000 (bilinear interpolation)
  - Input < 3000x3000: Scale down to 1400x1400 (RSS recommended size)
  - Input = 3000x3000: Use as-is
- Scaler: `golang.org/x/image/draw.BiLinear` (same as Jivefire thumbnail scaling)
- Maximum embedded size: ~500KB reasonable for PNG artwork

---

## Hugo Workflow Integration

### Episode Metadata Location

```
website/
└── content/
    └── episode/
        ├── 67.md          # Episode frontmatter
        ├── 68.md
        └── ...
```

### Frontmatter Structure

```yaml
---
title: "Mirrors, Motors and Makefiles"
podcast_duration: "54:09"
podcast_bytes: 27357184
# ... other fields
---
```

### Extraction Logic

```go
// Read episode file
path := filepath.Join(websiteRoot, "content", "episode", fmt.Sprintf("%d.md", episodeNum))
content, err := os.ReadFile(path)

// Extract title
re := regexp.MustCompile(`^title:\s*"(.+)"$`)
matches := re.FindSubmatch(content)
title := string(matches[1])
```

### Cover Art Location

```go
coverPath := filepath.Join(websiteRoot, "static", "img", "episode", "linuxmatters-3000x3000.png")
```

### Output Format

Display frontmatter-ready values:

```
Podcast frontmatter:
  podcast_duration: 54:09
  podcast_bytes: 27357184
```

User copies these into episode markdown.

---

## Error Handling Strategy

### Graceful Failures

1. **Missing episode file:**
```
Error: Episode file not found
  Expected: /path/to/website/content/episode/67.md

  Make sure episode markdown exists before encoding.
```

2. **Missing cover art:**
```
Error: Cover art not found
  Expected: /path/to/website/static/img/episode/linuxmatters-3000x3000.png

  Use --cover flag to specify custom artwork path.
```

3. **Encoding failure:**
```
Error: MP3 encoding failed

  This usually means:
  - Input file is corrupted
  - Unsupported audio format
  - Insufficient disk space

  Check input file with: ffprobe <input-file>
```

4. **ID3 writing failure:**
```
Error: Failed to write ID3 tags

  Partial file created: LMP67.mp3

  The MP3 is valid but missing metadata.
  You can re-run to add tags, or use --force to overwrite.
```

### Cleanup Strategy

- On encoding failure: delete partial MP3
- On ID3 failure: keep MP3, warn about missing metadata
- On overwrite prompt: only delete after user confirms

---

## Testing Strategy

### Unit Tests

```go
// Metadata extraction
func TestExtractEpisodeTitle(t *testing.T)
func TestExtractEpisodeTitleMissing(t *testing.T)
func TestExtractEpisodeTitleMalformed(t *testing.T)

// ID3 tag writing
func TestWriteID3Tags(t *testing.T)
func TestEmbedCoverArt(t *testing.T)

// File statistics
func TestCalculateDuration(t *testing.T)
func TestCalculateFileSize(t *testing.T)
```

### Integration Tests

```go
// Full encoding pipeline
func TestEncodeWAVToMP3(t *testing.T)
func TestEncodeFLACToMP3(t *testing.T)

// Workflow integration
func TestHugoWorkflow(t *testing.T)  // Mock website structure
```

### Test Data

```
testdata/
├── LMP0.flac              # 30-second clip
├── cover.png              # test artwork
└── 0.md                   # Mock episode file
```

---

## Performance Targets

### Encoding Speed
- Goal: >100x realtime on modern hardware
- Acceptable: >50x realtime
- MP3 encoding much faster than H.264 video (Jivefire)

### Memory Usage
- Peak: <100MB
- Acceptable: <200MB
- Streaming encode/decode (no full file buffering)

### Binary Size
- Goal: <70MB (slightly larger than Jivefire due to ID3 library)
- Acceptable: <100MB
- Static FFmpeg libraries are bulk of size

---

## Success Criteria

### MVP Complete When:

1. ✅ Encodes WAV/FLAC to MP3 matching `mp3-encode.sh` quality
2. ✅ Embeds all ID3v2 tags correctly
3. ✅ Embeds cover artwork
4. ✅ Extracts episode title from Hugo markdown
5. ✅ Displays podcast frontmatter values (duration, bytes)
6. ✅ Bubbletea UI shows encoding progress
7. ✅ Single binary with zero external dependencies
8. ✅ Works on Linux and macOS
9. ✅ Error handling matches or exceeds bash script
10. ✅ All tests passing

### Ready for Production When:

- Martin successfully uses it to encode an episode
- Output MP3 plays correctly in Apple Podcasts, Spotify, etc.
- Frontmatter values match expected format
- Cover art displays in podcast apps
- Tool feels as polished as Jivefire

---

## Open Questions

1. **ID3 version:** ID3v2.3 (widely compatible) or ID3v2.4 (more features)?
   - v2.3: Better compatibility with older players
   - v2.4: UTF-8 support, better chapter markers
   - Recommendation: v2.3 for broad compatibility

2. **Multiple presenters:** Should episode markdown support multiple artists (Mark, Martin, Popey)?
   - Single artist: "Linux Matters" (current approach)
   - Multiple: TPE1 (lead), TPE2 (album artist), TIPL (involved people)
   - Recommendation: keep simple initially, enhance later

---

## Development Timeline

**Week 1:** Foundation + Core Encoding
**Week 2:** Metadata + ID3 Tags
**Week 3:** Bubbletea UI + Statistics
**Week 4:** Hardening + Testing

**Total:** 4 weeks to production-ready MVP

---

## Conclusion

Jivedrop replaces `mp3-encode.sh` with a robust, cross-platform Go tool that:
- Eliminates fragile external dependencies
- Provides professional progress indication
- Integrates seamlessly with Linux Matters Hugo workflow
- Shares infrastructure with Jivefire and Jivetalking
- Delivers single-binary convenience

The tool maintains the "Jive" family aesthetic and engineering standards while solving a specific problem: podcast MP3 distribution encoding.

## Onboard

This a Jivedrop, a Go project, that encodes podcast audio file to MP3 (with covert art and ID3v2 tags) suitable for including in a podcast RSS feed.

Orientate yourself with the project by reading `README.md` and `docs/SPECIFICATION.md` and analysing the code. You should refer to the `ffmpeg-go` source code when required, it can usually be found in `/tmp/ffmpeg-go-research`, but if it is not there you can use `gh` to clone it from https://github.com/csnewman/ffmpeg-go

Sample data is in `testdata/`. You should only build and test via `just` commands. We are using NixOS as the host operating system and `flake.nix` provides tooling for the development shell. I use the `fish` shell. If you need to create "throw-away" test code, the put it in `testdata/`.

You are humble. You believe it is tasteless to claim your work is "Perfect", "Excellent" or "Production ready". The human collaborator will judge the quality of the work we do. Never claim something is fixed, working or implemented until the human collaborator has confirmed so.

Let me know when you are ready to start collaborating.
