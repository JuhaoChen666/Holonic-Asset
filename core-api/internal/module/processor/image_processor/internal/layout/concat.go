package layout

import (
	"fmt"
	"image"
	"image/draw"
)

type Direction uint8

const (
	Horizontal Direction = iota
	Vertical
)

func Concat(sources []*image.NRGBA, direction Direction, gap int) (*image.NRGBA, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("at least one image is required")
	}
	if gap < 0 {
		return nil, fmt.Errorf("gap cannot be negative")
	}

	width, height := dimensions(sources, direction, gap)
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	offset := 0
	for _, source := range sources {
		x, y := offset, (height-source.Bounds().Dy())/2
		if direction == Vertical {
			x, y = (width-source.Bounds().Dx())/2, offset
		}
		destination := image.Rect(
			x,
			y,
			x+source.Bounds().Dx(),
			y+source.Bounds().Dy(),
		)
		draw.Draw(result, destination, source, source.Bounds().Min, draw.Over)
		if direction == Horizontal {
			offset += source.Bounds().Dx() + gap
		} else {
			offset += source.Bounds().Dy() + gap
		}
	}
	return result, nil
}

func dimensions(sources []*image.NRGBA, direction Direction, gap int) (int, int) {
	width, height := 0, 0
	for _, source := range sources {
		if direction == Horizontal {
			width += source.Bounds().Dx()
			height = max(height, source.Bounds().Dy())
		} else {
			width = max(width, source.Bounds().Dx())
			height += source.Bounds().Dy()
		}
	}
	if direction == Horizontal {
		width += gap * (len(sources) - 1)
	} else {
		height += gap * (len(sources) - 1)
	}
	return width, height
}
