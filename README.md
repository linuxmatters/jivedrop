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
  - Episode title, number, album, artist, date, comment
  - Embedded cover artwork (PNG)
  - Podcast enclosure stats for duration and bytes
- ♊ **Dual-mode workflow**
  - 📝 **Hugo mode**—read metadata from episode markdown
  - 🎙️ **Standalone mode**—specify metadata via flags
- 🚀 **Single binary** Just drop and encode
  - 🐧 **Linux** (amd64 and aarch64)
  - 🍏 **macOS** (x86 and Apple Silicon)

### Example Output

<div align="center"><img alt="Jivedrop Demo" src=".github/jivedrop.gif" width="600" /></div>

## Usage

### Hugo Mode (Integrated Workflow)

For podcasts using Hugo static site generator and the something like [Castanet](https://github.com/mattstratton/castanet), Jivedrop reads metadata from episode markdown:

**Hugo mode automatically:**
- Reads episode title and number from frontmatter
- Locates cover art from `podcast_image` field
- Applies Linux Matters defaults (artist, album, comment)
- Outputs frontmatter-ready values for `podcast_duration` and `podcast_bytes`
- Prompts to update Hugo frontmatter

```bash
# Basic encoding
jivedrop LMP67.flac episode/67.md

# Override Hugo defaults
jivedrop LMP67.flac episode/67.md --artist "Ubuntu Podcast" --comment "https://ubuntupodcast.org"
```
### Standalone Mode (Universal Workflow)

**Standalone mode features:**
- Required flags: `--title`, `--num`, and `--cover`
- Optional metadata: `--artist`, `--album`, `--date`, `--comment`
- Smart filename generation: `{artist}-{num}.mp3` or `episode-{num}.mp3`
- Album defaults to artist value if not specified

For podcasts without Hugo—specify metadata via flags:

```bash
# Minimal (title, episode number, and cover art required)
jivedrop audio.flac \
  --title "Terminal Full of Sparkles" \
  --num 66 \
  --cover artwork.png

# Full metadata
jivedrop audio.flac \
  --title "Terminal Full of Sparkles" \
  --num 66 \
  --artist "Linux Matters" \
  --album "All Seasons" \
  --date "2025-10" \
  --comment "https://linuxmatters.sh/66" \
  --cover artwork.png
```

## CLI Reference

```
Usage:
  Hugo mode:
    jivedrop <audio-file> <episode-md> [flags]
  Standalone mode:
    jivedrop <audio-file> --title TEXT --num NUMBER --cover PATH [flags]


Arguments:
  [<audio-file>]  Path to audio file (WAV, FLAC)
  [<episode-md>]  Path to episode markdown file (Hugo mode)


Flags:
  -h, --help  Show context-sensitive help.
  --title     Episode title (required in standalone mode)
  --num       Episode number (required in standalone mode)
  --cover     Cover art path (required in standalone mode)
  --artist    Artist name (defaults to 'Linux Matters' in Hugo mode)
  --album     Album name (defaults to artist value if omitted)
  --date      Release date (YYYY-MM-DD format)
  --comment   Comment URL (defaults to 'https://linuxmatters.sh' in Hugo mode)
  --output-path  Output file or directory path (default: STRING)
  --stereo  Encode as stereo at 192kbps (default: mono at 112kbps) (default: BOOL)
  --version  Show version information (default: BOOL)
```

### Output
- Hugo mode:        `LMP{num}.mp3` (or `{artist}-{num}.mp3` with `--artist` override)
- Standalone mode:  `{artist}-{num}.mp3` (or `episode-{num}.mp3` without `--artist`)

### Encoding settings
- Mono:   112kbps CBR, 44.1kHz, quality 3, 20.5kHz lowpass
- Stereo: 192kbps CBR, 44.1kHz, quality 3, 20.5kHz lowpass

### ID3v2 Tags
- `TIT2`: `{num}: {title}`
- `TALB`: `{album}` (omitted if not provided)
- `TRCK`: `{num}`
- `TPE1`: `{artist}` (omitted if not provided)
- `TDRC`: `{date}` (defaults to current YYYY-MM)
- `COMM`: `{comment}` (omitted if not provided)
- `APIC`: Cover art (PNG, front cover)

## Build

```bash
just build      # Build binary
just mp3        # Encode test audio

# Manual
go build -o jivedrop ./cmd/jivedrop
```

## Specification

The complete Jivedrop specification is available in [SPECIFICATION.md](docs/SPECIFICATION.md).
