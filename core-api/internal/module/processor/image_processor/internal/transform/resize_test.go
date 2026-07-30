package transform

import (
	"errors"
	"image"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
)

func TestResizeRejectsHugeOutputWithoutPanic(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	_, err := Resize(source, int(^uint(0)>>1), 2, PixelArt)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Resize error = %v, want ErrResourceLimit", err)
	}
}

func TestResizeAllowsTilesetDimensions(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	result, err := Resize(source, 32, 64, PixelArt)
	if err != nil {
		t.Fatalf("Resize returned an error: %v", err)
	}
	if got := result.Bounds().Size(); got != image.Pt(32, 64) {
		t.Fatalf("Resize size = %v, want (32,64)", got)
	}
}
