package id3

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/linuxmatters/jivedrop/internal/cli"
	"golang.org/x/image/draw"
)

// ScaleCoverArt scales cover art according to Apple Podcasts specifications:
//   - Images < 1400x1400: upscale to 1400x1400
//   - Images 1400x1400 to 3000x3000: use as-is (no scaling artifacts)
//   - Images > 3000x3000: downscale to 3000x3000
//
// This preserves quality by avoiding unnecessary scaling when images are
// already within the acceptable range (1400-3000).
func ScaleCoverArt(inputPath string) ([]byte, error) {
	// Open and decode the image
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cover art: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cover art: %w", err)
	}

	// Verify image is square
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != height {
		return nil, fmt.Errorf("cover art must be square (got %dx%d)", width, height)
	}

	// Determine target size based on current dimensions
	var targetSize int
	needsScaling := false

	switch {
	case width < 1400:
		// Too small: upscale to 1400
		targetSize = 1400
		needsScaling = true
	case width >= 1400 && width <= 3000:
		// Perfect range: use as-is
		targetSize = width
		needsScaling = false
	case width > 3000:
		// Too large: downscale to 3000
		targetSize = 3000
		needsScaling = true
	}

	// Scale if needed
	var finalImg image.Image
	if needsScaling {
		// Create target image
		dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))

		// Use bilinear interpolation for high-quality scaling
		// Same quality scaler as Jivefire thumbnail generation
		draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

		finalImg = dst
	} else {
		// Use original image unchanged
		finalImg = img
	}

	// Re-encode to PNG
	// Note: We always use PNG regardless of input format to ensure
	// consistent quality in ID3 tags
	var buf []byte
	pngBuf := &bytesBuffer{buf: &buf}

	err = png.Encode(pngBuf, finalImg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode scaled image: %w", err)
	}

	// Log scaling decision
	if needsScaling {
		cli.PrintCover(fmt.Sprintf("%dx%d %s scaled to %dx%d PNG", width, height, format, targetSize, targetSize))
	} else {
		cli.PrintCover(fmt.Sprintf("%dx%d %s (no scaling needed)", width, height, format))
	}

	return buf, nil
}

// bytesBuffer implements io.Writer for capturing PNG encoding output
type bytesBuffer struct {
	buf *[]byte
}

func (b *bytesBuffer) Write(p []byte) (n int, err error) {
	*b.buf = append(*b.buf, p...)
	return len(p), nil
}
