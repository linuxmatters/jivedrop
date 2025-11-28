# Jivedrop - Just Commands
# Drop your podcast audio into RSS-ready MP3s

# Default recipe (shows available commands)
default:
    @just --list

# Download ffmpeg-statigo libraries and configure git for submodules
setup:
    #!/usr/bin/env bash
    echo "Configuring git for submodule-friendly pulls..."
    git config pull.ff only
    git config submodule.recurse true
    echo "Initialising ffmpeg-statigo submodule..."
    git submodule update --init --recursive
    echo "Downloading ffmpeg-statigo libraries..."
    cd third_party/ffmpeg-statigo && go run ./cmd/download-lib
    echo "Setup complete!"

# Update ffmpeg-statigo submodule
update-ffmpeg:
    #!/usr/bin/env bash
    echo "Updating ffmpeg-statigo submodule..."
    cd third_party/ffmpeg-statigo
    git pull origin main
    cd ../..
    git add third_party/ffmpeg-statigo
    echo "Submodule updated"
    just setup
    echo "Don't forget to commit: git commit -m 'chore: update ffmpeg-statigo submodule'"

# Build jivedrop (dev version)
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

# Run MP3 encoding test
mp3: build
    @echo n | ./jivedrop testdata/LMP67.flac testdata/67.md --output-path testdata/

# Record gif
vhs: build
    @vhs ./jivedrop.tape
    rm LMP67.mp3 2>/dev/null || true

# Run tests
test:
    go test ./...
