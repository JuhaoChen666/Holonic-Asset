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

func Resize(
	source *image.NRGBA,
	width int,
	height int,
	filter ResizeFilter,
) (*image.NRGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("resize dimensions must be positive")
	}
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	scaler := xdraw.Scaler(xdraw.CatmullRom)
	if filter == PixelArt {
		scaler = xdraw.NearestNeighbor
	}
	scaler.Scale(result, result.Bounds(), source, source.Bounds(), draw.Src, nil)
	return result, nil
}
