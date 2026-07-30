package transform

import (
	"fmt"
	"image"
	"image/draw"

	xdraw "golang.org/x/image/draw"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
)

type ResizeFilter uint8

const (
	Smooth ResizeFilter = iota
	PixelArt
)

func Resize(
	source *image.NRGBA,
	width int,
	height int,
	filter ResizeFilter,
) (*image.NRGBA, error) {
	if source == nil {
		return nil, fmt.Errorf("resize source is required")
	}
	sourcePixels, err := limits.PixelCount(
		"resize source",
		source.Bounds().Dx(),
		source.Bounds().Dy(),
		limits.MaxImagePixels,
	)
	if err != nil {
		return nil, err
	}
	outputPixels, err := limits.PixelCount(
		"resize output",
		width,
		height,
		limits.MaxOutputPixels,
	)
	if err != nil {
		return nil, err
	}
	workingPixels, err := limits.CheckedAdd(
		"resize working pixels",
		sourcePixels,
		outputPixels,
	)
	if err != nil {
		return nil, err
	}
	if err := limits.CheckMaximum(
		"resize working pixels",
		workingPixels,
		limits.MaxWorkingPixels,
	); err != nil {
		return nil, err
	}
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	scaler := xdraw.Scaler(xdraw.CatmullRom)
	if filter == PixelArt {
		scaler = xdraw.NearestNeighbor
	}
	scaler.Scale(result, result.Bounds(), source, source.Bounds(), draw.Src, nil)
	return result, nil
}
