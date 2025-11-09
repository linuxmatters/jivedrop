# Jivedrop - Just Commands
# Drop your podcast audio into RSS-ready MP3s

# Default recipe (shows available commands)
default:
    @just --list

# Build the jivedrop binary (dev version)
build:
    #!/usr/bin/env bash
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    echo "Building jivedrop version: $VERSION"
    CGO_ENABLED=1 go build -ldflags="-X main.version=$VERSION" -o jivedrop ./cmd/jivedrop

# Clean build artifacts
clean:
    rm -fv jivedrop 2>/dev/null || true
    @rm testdata/*.mp3 2>/dev/null || true

# Install the jivedrop binary to ~/.local/bin
install: build
    @mkdir -p ~/.local/bin 2>/dev/null || true
    @mv ./jivedrop ~/.local/bin/jivedrop
    @echo "Installed jivedrop to ~/.local/bin/jivedrop"
    @echo "Make sure ~/.local/bin is in your PATH"

# Run MP3 encodinging test
mp3: build
    @echo n | ./jivedrop testdata/0.md testdata/LMP0.flac testdata/

vhs: build
    @vhs ./jivedrop.tape

# Run tests
test:
    go test ./...

# Get project orientation info
onboard:
  @cat docs/SPECIFICATION.md | grep -A 20 "^## Onboard"
