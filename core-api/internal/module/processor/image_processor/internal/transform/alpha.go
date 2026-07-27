package transform

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func RemoveBackground(
	source *image.NRGBA,
	background *color.NRGBA,
	tolerance uint8,
	feather uint8,
) *image.NRGBA {
	result := clone(source)
	backgroundColor := cornerAverage(result)
	if background != nil {
		backgroundColor = *background
	}
	for y := 0; y < result.Bounds().Dy(); y++ {
		for x := 0; x < result.Bounds().Dx(); x++ {
			pixel := result.NRGBAAt(x, y)
			distance := colorDistance(pixel, backgroundColor)
			switch {
			case distance <= int(tolerance):
				pixel.A = 0
			case feather > 0 && distance < int(tolerance)+int(feather):
				ratio := float64(distance-int(tolerance)) / float64(feather)
				pixel.A = uint8(math.Round(float64(pixel.A) * ratio))
			}
			result.SetNRGBA(x, y, pixel)
		}
	}
	return result
}

func TrimTransparent(source *image.NRGBA, alphaThreshold uint8) (*image.NRGBA, error) {
	bounds, ok := alphaBounds(source, alphaThreshold)
	if !ok {
		return nil, fmt.Errorf("image has no pixels above the alpha threshold")
	}
	result := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)
	return result, nil
}

func alphaBounds(source *image.NRGBA, threshold uint8) (image.Rectangle, bool) {
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	left, top, right, bottom := width, height, 0, 0
	found := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if source.NRGBAAt(x, y).A <= threshold {
				continue
			}
			found = true
			left = min(left, x)
			top = min(top, y)
			right = max(right, x+1)
			bottom = max(bottom, y+1)
		}
	}
	return image.Rect(left, top, right, bottom), found
}

func cornerAverage(source *image.NRGBA) color.NRGBA {
	maxX := source.Bounds().Dx() - 1
	maxY := source.Bounds().Dy() - 1
	corners := [...]color.NRGBA{
		source.NRGBAAt(0, 0),
		source.NRGBAAt(maxX, 0),
		source.NRGBAAt(0, maxY),
		source.NRGBAAt(maxX, maxY),
	}
	var red, green, blue int
	for _, corner := range corners {
		red += int(corner.R)
		green += int(corner.G)
		blue += int(corner.B)
	}
	return color.NRGBA{
		R: uint8(red / len(corners)),
		G: uint8(green / len(corners)),
		B: uint8(blue / len(corners)),
		A: 255,
	}
}

func colorDistance(left color.NRGBA, right color.NRGBA) int {
	return max(
		abs(int(left.R)-int(right.R)),
		abs(int(left.G)-int(right.G)),
		abs(int(left.B)-int(right.B)),
	)
}

func clone(source *image.NRGBA) *image.NRGBA {
	result := image.NewNRGBA(image.Rect(0, 0, source.Bounds().Dx(), source.Bounds().Dy()))
	draw.Draw(result, result.Bounds(), source, source.Bounds().Min, draw.Src)
	return result
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
