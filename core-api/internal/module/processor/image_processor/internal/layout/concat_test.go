package layout

import (
	"errors"
	"image"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
)

func TestDimensions(t *testing.T) {
	t.Parallel()

	width, height, sourcePixels, err := Dimensions(
		[]image.Point{image.Pt(32, 64), image.Pt(64, 96)},
		Vertical,
		0,
	)
	if err != nil {
		t.Fatalf("Dimensions returned an error: %v", err)
	}
	if width != 64 || height != 160 || sourcePixels != 8192 {
		t.Fatalf(
			"Dimensions = %dx%d with %d source pixels, want 64x160 with 8192",
			width,
			height,
			sourcePixels,
		)
	}
}

func TestDimensionsRejectsHugeGapWithoutPanic(t *testing.T) {
	t.Parallel()

	_, _, _, err := Dimensions(
		[]image.Point{image.Pt(1, 1), image.Pt(1, 1)},
		Horizontal,
		int(^uint(0)>>1),
	)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Dimensions error = %v, want ErrResourceLimit", err)
	}
}

func TestDimensionsRejectsTooManyInputs(t *testing.T) {
	t.Parallel()

	sizes := make([]image.Point, limits.MaxConcatInputs+1)
	for index := range sizes {
		sizes[index] = image.Pt(1, 1)
	}
	_, _, _, err := Dimensions(sizes, Horizontal, 0)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Dimensions error = %v, want ErrResourceLimit", err)
	}
}
