# Jivedrop 🪩

> Drop your podcast .wav into distribution-ready MP3. ID3 tags, cover art, and frontmatter values—all in one go.

## The Groove

Your podcast deserves proper packaging. Jivedrop transforms mixed audio (WAV/FLAC) into RSS-ready MP3 files with embedded artwork, complete metadata, and zero hassle.

### What's Cooking

- 🎵 **CBR 112kbps mono MP3**—optimized for podcast distribution
  - 🎚️ **44.1kHz resampling** with automatic mono downmix
  - 🎼 **20.5kHz lowpass filter** for clean high-frequency rolloff
  - 🏆 **LAME quality preset 3**—excellent quality, reasonable file size
- 🏷️ **Complete ID3v2 metadata**
  - Episode title, number, album, artist, date, URL
  - Embedded cover artwork (PNG)
  - Hugo frontmatter values (duration, bytes)
- 🚀 **Single binary** Just drop and encode
  - 🐧 **Linux** (amd64 and aarch64)
  - 🍏 **macOS** (x86 and Apple Silicon)

## Usage

### Basic Encoding
```bash
./jivedrop /path/to/epidode.md /path/to/LMP67.wav
```

### Custom Output Directory
```bash
./jivedrop /path/to/epidode.md /path/to/LMP67.wav /tmp
```

### Advanced Options
```bash
# Stereo encoding at 192kbps
./jivedrop --stereo --bitrate=192 /path/to/epidode.md input.wav

# Custom cover art
./jivedrop --cover=/path/to/artwork.png /path/to/epidode.md input.flac
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
  Date:    2025
  Comment: https://linuxmatters.sh/
  Cover:   3000x3000 PNG (482KB)

✓ Complete: LMP67.mp3

Podcast frontmatter:
  podcast_duration: 54:09
  podcast_bytes: 27357184
```

## Hugo Workflow Integration

Jivedrop automatically:
- Reads episode title from `episode.md`
- Locates cover art at from frontmatter.`
- Outputs frontmatter-ready values for your episode markdown

Drop the `podcast_duration` and `podcast_bytes` values straight into your Hugo frontmatter.

## Build

```bash
just build      # Build binary
just test       # Encode test audio

# Manual
go build -o jivedrop ./cmd/jivedrop
```

## CLI Reference

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

## Specification

The complete Jivedrop specification is available in [SPECIFICATION.md](docs/SPECIFICATION.md).
