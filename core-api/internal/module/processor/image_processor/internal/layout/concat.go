package layout

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
)

type Direction uint8

const (
	Horizontal Direction = iota
	Vertical
)

func Concat(sources []*image.NRGBA, direction Direction, gap int) (*image.NRGBA, error) {
	sizes := make([]image.Point, 0, len(sources))
	for index, source := range sources {
		if source == nil {
			return nil, fmt.Errorf("image %d is required", index)
		}
		sizes = append(sizes, source.Bounds().Size())
	}

	width, height, _, err := Dimensions(sizes, direction, gap)
	if err != nil {
		return nil, err
	}
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

// Dimensions preflights a Concat canvas before decoded images are allocated.
func Dimensions(sizes []image.Point, direction Direction, gap int) (int, int, int, error) {
	if len(sizes) == 0 {
		return 0, 0, 0, fmt.Errorf("at least one image is required")
	}
	if direction != Horizontal && direction != Vertical {
		return 0, 0, 0, fmt.Errorf("unsupported concat direction: %d", direction)
	}
	if err := limits.CheckMaximum(
		"concat input count",
		len(sizes),
		limits.MaxConcatInputs,
	); err != nil {
		return 0, 0, 0, err
	}
	if gap < 0 {
		return 0, 0, 0, fmt.Errorf("gap cannot be negative")
	}

	width, height, sourcePixels := 0, 0, 0
	for index, size := range sizes {
		pixels, err := limits.PixelCount(
			fmt.Sprintf("concat image %d", index),
			size.X,
			size.Y,
			limits.MaxImagePixels,
		)
		if err != nil {
			return 0, 0, 0, err
		}
		sourcePixels, err = limits.CheckedAdd(
			"concat source pixels",
			sourcePixels,
			pixels,
		)
		if err != nil {
			return 0, 0, 0, err
		}
		if direction == Horizontal {
			width, err = limits.CheckedAdd("concat width", width, size.X)
			height = max(height, size.Y)
		} else {
			width = max(width, size.X)
			height, err = limits.CheckedAdd("concat height", height, size.Y)
		}
		if err != nil {
			return 0, 0, 0, err
		}
	}

	totalGap, err := limits.CheckedMultiply("concat gap", gap, len(sizes)-1)
	if err != nil {
		return 0, 0, 0, err
	}
	if direction == Horizontal {
		width, err = limits.CheckedAdd("concat width", width, totalGap)
	} else {
		height, err = limits.CheckedAdd("concat height", height, totalGap)
	}
	if err != nil {
		return 0, 0, 0, err
	}

	outputPixels, err := limits.PixelCount(
		"concat output",
		width,
		height,
		limits.MaxOutputPixels,
	)
	if err != nil {
		return 0, 0, 0, err
	}
	workingPixels, err := limits.CheckedAdd(
		"concat working pixels",
		sourcePixels,
		outputPixels,
	)
	if err != nil {
		return 0, 0, 0, err
	}
	if err := limits.CheckMaximum(
		"concat working pixels",
		workingPixels,
		limits.MaxWorkingPixels,
	); err != nil {
		return 0, 0, 0, err
	}
	return width, height, sourcePixels, nil
}
