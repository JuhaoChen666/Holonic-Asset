package transform

import (
	"fmt"
	"image"
	"image/draw"

	xdraw "golang.org/x/image/draw"
)

type ResizeFilter uint8

const (
	Smooth ResizeFilter = iota
	PixelArt
)

type ResizeAnchor uint8

const (
	AnchorCenter ResizeAnchor = iota
	AnchorBottomCenter
)

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
	anchor ResizeAnchor,
) (*image.NRGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("resize dimensions must be positive")
	}
	if source == nil || source.Bounds().Empty() {
		return nil, fmt.Errorf("resize source must not be empty")
	}
	scaler, err := scalerFor(filter)
	if err != nil {
		return nil, err
	}
	if anchor != AnchorCenter && anchor != AnchorBottomCenter {
		return nil, fmt.Errorf("unsupported resize anchor: %d", anchor)
	}

	sourceWidth := source.Bounds().Dx()
	sourceHeight := source.Bounds().Dy()
	scaledWidth, scaledHeight := containDimensions(
		sourceWidth,
		sourceHeight,
		width,
		height,
	)
	x := (width - scaledWidth) / 2
	y := (height - scaledHeight) / 2
	if anchor == AnchorBottomCenter {
		y = height - scaledHeight
	}

	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	destination := image.Rect(x, y, x+scaledWidth, y+scaledHeight)
	scaler.Scale(result, destination, source, source.Bounds(), draw.Src, nil)
	return result, nil
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
