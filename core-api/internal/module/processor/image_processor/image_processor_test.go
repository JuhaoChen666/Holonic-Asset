package imageprocessor

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"testing"
)

func TestCropRejectsNegativeDimensionsBeforeDecoding(t *testing.T) {
	t.Parallel()

	_, err := New().Crop(ImageInput{}, Rect{X: 8, Y: 0, Width: -3, Height: 1})
	if err == nil || !strings.Contains(err.Error(), "dimensions must be positive") {
		t.Fatalf("Crop() error = %v, want positive-dimensions error", err)
	}
}

func TestResizeRejectsUnknownFilterBeforeDecoding(t *testing.T) {
	t.Parallel()

	_, err := New().Resize(ImageInput{}, 1, 1, ResizeFilter(255))
	if err == nil || !strings.Contains(err.Error(), "unsupported resize filter") {
		t.Fatalf("Resize() error = %v, want unsupported-filter error", err)
	}
}

func TestConcatRejectsUnknownDirectionBeforeDecoding(t *testing.T) {
	t.Parallel()

	_, err := New().Concat(
		[]ImageInput{{Base64: "not-an-image"}},
		ConcatDirection(255),
		0,
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported concat direction") {
		t.Fatalf("Concat() error = %v, want unsupported-direction error", err)
	}
}

func TestSetOpacityRejectsNonFiniteValuesBeforeDecoding(t *testing.T) {
	t.Parallel()

	for _, opacity := range []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	} {
		_, err := New().SetOpacity(ImageInput{}, opacity)
		if err == nil || !strings.Contains(err.Error(), "finite number") {
			t.Errorf("SetOpacity(%v) error = %v, want finite-number error", opacity, err)
		}
	}
}

func TestAdjustRGBAPreservesFullyTransparentPixels(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	source.SetNRGBA(0, 0, color.NRGBA{R: 20, G: 30, B: 40, A: 0})
	source.SetNRGBA(1, 0, color.NRGBA{R: 20, G: 30, B: 40, A: 20})

	output, err := New().AdjustRGBA(
		testImageInput(t, source),
		RGBAAdjustment{
			Alpha:               32,
			PreserveTransparent: true,
		},
	)
	if err != nil {
		t.Fatalf("AdjustRGBA() error = %v", err)
	}
	result := decodeTestOutput(t, output)
	if alpha := result.NRGBAAt(0, 0).A; alpha != 0 {
		t.Errorf("transparent pixel alpha = %d, want 0", alpha)
	}
	if alpha := result.NRGBAAt(1, 0).A; alpha != 52 {
		t.Errorf("visible pixel alpha = %d, want 52", alpha)
	}
}

func TestAdjustRGBAChangesTransparentPixelsWhenNotPreserved(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	source.SetNRGBA(0, 0, color.NRGBA{R: 20, G: 30, B: 40, A: 0})

	output, err := New().AdjustRGBA(
		testImageInput(t, source),
		RGBAAdjustment{Alpha: 32},
	)
	if err != nil {
		t.Fatalf("AdjustRGBA() error = %v", err)
	}
	result := decodeTestOutput(t, output)
	if alpha := result.NRGBAAt(0, 0).A; alpha != 32 {
		t.Errorf("transparent pixel alpha = %d, want 32", alpha)
	}
}

func TestRemoveBackgroundUsesExplicitToleranceAndFeather(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	background := color.NRGBA{R: 237, G: 16, B: 255, A: 255}
	for y := range 3 {
		for x := range 3 {
			source.SetNRGBA(x, y, background)
		}
	}
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 255, A: 255})
	source.SetNRGBA(2, 0, color.NRGBA{R: 219, G: 0, B: 255, A: 255})
	source.SetNRGBA(0, 2, color.NRGBA{R: 255, G: 32, B: 255, A: 255})
	source.SetNRGBA(2, 2, color.NRGBA{R: 219, G: 32, B: 255, A: 255})
	source.SetNRGBA(1, 1, color.NRGBA{R: 20, G: 30, B: 40, A: 255})
	input := testImageInput(t, source)
	backgroundColor := &RGB{Red: 255, Green: 0, Blue: 255}

	exactOutput, err := New().RemoveBackground(input, RemoveBackgroundOptions{
		Background: backgroundColor,
	})
	if err != nil {
		t.Fatalf("RemoveBackground(exact) error = %v", err)
	}
	thresholdOutput, err := New().RemoveBackground(input, RemoveBackgroundOptions{
		Background: backgroundColor,
		Tolerance:  40,
		Feather:    4,
	})
	if err != nil {
		t.Fatalf("RemoveBackground(threshold) error = %v", err)
	}
	if exactOutput.Base64 == thresholdOutput.Base64 {
		t.Fatal("explicit threshold settings unexpectedly produced identical output")
	}

	exactResult := decodeTestOutput(t, exactOutput)
	if alpha := exactResult.NRGBAAt(2, 0).A; alpha != 255 {
		t.Errorf("exact-match corner alpha = %d, want 255", alpha)
	}

	thresholdResult := decodeTestOutput(t, thresholdOutput)
	for _, point := range []image.Point{
		{X: 0, Y: 0},
		{X: 2, Y: 0},
		{X: 0, Y: 2},
		{X: 2, Y: 2},
	} {
		if alpha := thresholdResult.NRGBAAt(point.X, point.Y).A; alpha != 0 {
			t.Errorf("corner %v alpha = %d, want 0", point, alpha)
		}
	}
	if alpha := thresholdResult.NRGBAAt(1, 1).A; alpha != 255 {
		t.Errorf("foreground alpha = %d, want 255", alpha)
	}
}

func TestResizeContainPreservesAspectRatioAndBottomAnchor(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := range 2 {
		for x := range 4 {
			source.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}

	output, err := New().ResizeContain(
		testImageInput(t, source),
		6,
		6,
		ResizePixelArt,
		ResizeAnchorBottomCenter,
	)
	if err != nil {
		t.Fatalf("ResizeContain() error = %v", err)
	}
	if output.Width != 6 || output.Height != 6 {
		t.Fatalf("ResizeContain() size = %dx%d, want 6x6", output.Width, output.Height)
	}

	result := decodeTestOutput(t, output)
	for y := range 3 {
		if alpha := result.NRGBAAt(0, y).A; alpha != 0 {
			t.Errorf("top padding alpha at y=%d is %d, want 0", y, alpha)
		}
	}
	for y := 3; y < 6; y++ {
		if alpha := result.NRGBAAt(0, y).A; alpha != 255 {
			t.Errorf("scaled source alpha at y=%d is %d, want 255", y, alpha)
		}
	}
}

func TestTrimTransparentReturnsSourceOffset(t *testing.T) {
	t.Parallel()

	source := image.NewNRGBA(image.Rect(0, 0, 5, 6))
	for y := 3; y < 5; y++ {
		for x := 2; x < 4; x++ {
			source.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}

	output, err := New().TrimTransparent(testImageInput(t, source), 0)
	if err != nil {
		t.Fatalf("TrimTransparent() error = %v", err)
	}
	if output.Width != 2 || output.Height != 2 {
		t.Errorf("trimmed size = %dx%d, want 2x2", output.Width, output.Height)
	}
	if output.OffsetX != 2 || output.OffsetY != 3 {
		t.Errorf(
			"trimmed offset = (%d,%d), want (2,3)",
			output.OffsetX,
			output.OffsetY,
		)
	}
}

func testImageInput(t *testing.T, source image.Image) ImageInput {
	t.Helper()

	var content bytes.Buffer
	if err := png.Encode(&content, source); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return ImageInput{Base64: base64.StdEncoding.EncodeToString(content.Bytes())}
}

func decodeTestOutput(t *testing.T, output ImageOutput) *image.NRGBA {
	t.Helper()

	content, err := base64.StdEncoding.DecodeString(output.Base64)
	if err != nil {
		t.Fatalf("decode output Base64: %v", err)
	}
	decoded, err := png.Decode(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("decode output PNG: %v", err)
	}
	result, ok := decoded.(*image.NRGBA)
	if !ok {
		t.Fatalf("output type = %T, want *image.NRGBA", decoded)
	}
	return result
}
