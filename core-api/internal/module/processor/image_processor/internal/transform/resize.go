package transform

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

type ResizeFilter uint8

const (
	Smooth ResizeFilter = iota
	PixelArt
)

type Insets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

func Resize(
	source *image.NRGBA,
	width int,
	height int,
	filter ResizeFilter,
) (*image.NRGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("resize dimensions must be positive")
	}
	scaler, err := scalerFor(filter)
	if err != nil {
		return nil, err
	}
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	scaler.Scale(result, result.Bounds(), source, source.Bounds(), draw.Src, nil)
	return result, nil
}

func ResizeContain(
	source *image.NRGBA,
	width int,
	height int,
	filter ResizeFilter,
	padding Insets,
	anchorX float64,
	anchorY float64,
	allowUpscale bool,
) (*image.NRGBA, image.Rectangle, error) {
	if width <= 0 || height <= 0 {
		return nil, image.Rectangle{}, fmt.Errorf("resize dimensions must be positive")
	}
	if source == nil || source.Bounds().Empty() {
		return nil, image.Rectangle{}, fmt.Errorf("resize source must not be empty")
	}
	scaler, err := scalerFor(filter)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	if padding.Top < 0 || padding.Right < 0 ||
		padding.Bottom < 0 || padding.Left < 0 {
		return nil, image.Rectangle{}, fmt.Errorf("resize padding cannot be negative")
	}
	availableWidth := width - padding.Left - padding.Right
	availableHeight := height - padding.Top - padding.Bottom
	if availableWidth <= 0 || availableHeight <= 0 {
		return nil, image.Rectangle{}, fmt.Errorf(
			"resize padding must leave a positive content area",
		)
	}
	if math.IsNaN(anchorX) || math.IsNaN(anchorY) ||
		math.IsInf(anchorX, 0) || math.IsInf(anchorY, 0) ||
		anchorX < 0 || anchorX > 1 || anchorY < 0 || anchorY > 1 {
		return nil, image.Rectangle{}, fmt.Errorf(
			"resize anchor coordinates must be finite numbers between 0 and 1",
		)
	}

	sourceWidth := source.Bounds().Dx()
	sourceHeight := source.Bounds().Dy()
	scaledWidth, scaledHeight := sourceWidth, sourceHeight
	if allowUpscale ||
		sourceWidth > availableWidth ||
		sourceHeight > availableHeight {
		scaledWidth, scaledHeight = containDimensions(
			sourceWidth,
			sourceHeight,
			availableWidth,
			availableHeight,
		)
	}
	freeX := availableWidth - scaledWidth
	freeY := availableHeight - scaledHeight
	x := padding.Left + int(math.Floor(float64(freeX)*anchorX))
	y := padding.Top + int(math.Floor(float64(freeY)*anchorY))

	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	destination := image.Rect(x, y, x+scaledWidth, y+scaledHeight)
	scaler.Scale(result, destination, source, source.Bounds(), draw.Src, nil)
	scrubTransparentRGB(result)
	return result, destination, nil
}

func containDimensions(
	sourceWidth int,
	sourceHeight int,
	targetWidth int,
	targetHeight int,
) (int, int) {
	if int64(sourceWidth)*int64(targetHeight) >
		int64(sourceHeight)*int64(targetWidth) {
		scaledHeight := roundedScale(sourceHeight, targetWidth, sourceWidth)
		return targetWidth, max(1, min(targetHeight, scaledHeight))
	}
	scaledWidth := roundedScale(sourceWidth, targetHeight, sourceHeight)
	return max(1, min(targetWidth, scaledWidth)), targetHeight
}

func roundedScale(value int, numerator int, denominator int) int {
	scaled := int64(value) * int64(numerator)
	return int((scaled + int64(denominator)/2) / int64(denominator))
}

func scalerFor(filter ResizeFilter) (xdraw.Scaler, error) {
	switch filter {
	case Smooth:
		return xdraw.CatmullRom, nil
	case PixelArt:
		return xdraw.NearestNeighbor, nil
	default:
		return nil, fmt.Errorf("unsupported resize filter: %d", filter)
	}
}
