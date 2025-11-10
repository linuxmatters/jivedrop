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
./jivedrop --stereo /path/to/epidode.md input.wav

# Custom cover art
./jivedrop --cover=/path/to/artwork.png /path/to/epidode.md input.flac
```

### Example Output

<div align="center"><img alt="Jivedrop Demo" src=".github/jivedrop.gif" width="600" /></div>

## Hugo Workflow Integration

Jivedrop automatically:
- Reads episode title from `episode.md`
- Locates cover art at from frontmatter
- Outputs frontmatter-ready values for `podcast_duration` and `podcast_bytes` and will prompt to automatically update your Hugo frontmatter

## Build

```bash
just build      # Build binary
just mp3        # Encode test audio

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
