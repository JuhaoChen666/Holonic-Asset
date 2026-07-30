package transform

import (
	"fmt"
	"image"
	"math"
)

func AdjustRGBA(
	source *image.NRGBA,
	red int,
	green int,
	blue int,
	alpha int,
	preserveTransparent bool,
) *image.NRGBA {
	result := clone(source)
	width, height := result.Bounds().Dx(), result.Bounds().Dy()
	for y := range height {
		for x := range width {
			pixel := result.NRGBAAt(x, y)
			wasTransparent := pixel.A == 0
			pixel.R = clampChannel(int(pixel.R) + red)
			pixel.G = clampChannel(int(pixel.G) + green)
			pixel.B = clampChannel(int(pixel.B) + blue)
			if !preserveTransparent || !wasTransparent {
				pixel.A = clampChannel(int(pixel.A) + alpha)
			}
			result.SetNRGBA(x, y, pixel)
		}
	}
	return result
}

func SetOpacity(source *image.NRGBA, opacity float64) (*image.NRGBA, error) {
	if math.IsNaN(opacity) || math.IsInf(opacity, 0) || opacity < 0 || opacity > 1 {
		return nil, fmt.Errorf("opacity must be a finite number between 0 and 1")
	}
	result := clone(source)
	width, height := result.Bounds().Dx(), result.Bounds().Dy()
	for y := range height {
		for x := range width {
			pixel := result.NRGBAAt(x, y)
			pixel.A = uint8(math.Round(float64(pixel.A) * opacity))
			result.SetNRGBA(x, y, pixel)
		}
	}
	return result, nil
}

func clampChannel(value int) uint8 {
	if value <= 0 {
		return 0
	}
	if value >= 255 {
		return 255
	}
	return uint8(value)
}
