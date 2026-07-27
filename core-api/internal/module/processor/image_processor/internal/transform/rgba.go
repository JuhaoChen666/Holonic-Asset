package transform

import (
	"fmt"
	"image"
	"math"
)

func AdjustRGBA(source *image.NRGBA, red int, green int, blue int, alpha int) *image.NRGBA {
	result := clone(source)
	for y := 0; y < result.Bounds().Dy(); y++ {
		for x := 0; x < result.Bounds().Dx(); x++ {
			pixel := result.NRGBAAt(x, y)
			pixel.R = clampChannel(int(pixel.R) + red)
			pixel.G = clampChannel(int(pixel.G) + green)
			pixel.B = clampChannel(int(pixel.B) + blue)
			pixel.A = clampChannel(int(pixel.A) + alpha)
			result.SetNRGBA(x, y, pixel)
		}
	}
	return result
}

func SetOpacity(source *image.NRGBA, opacity float64) (*image.NRGBA, error) {
	if opacity < 0 || opacity > 1 {
		return nil, fmt.Errorf("opacity must be between 0 and 1")
	}
	result := clone(source)
	for y := 0; y < result.Bounds().Dy(); y++ {
		for x := 0; x < result.Bounds().Dx(); x++ {
			pixel := result.NRGBAAt(x, y)
			pixel.A = uint8(math.Round(float64(pixel.A) * opacity))
			result.SetNRGBA(x, y, pixel)
		}
	}
	return result, nil
}

func clampChannel(value int) uint8 {
	return uint8(max(0, min(255, value)))
}
