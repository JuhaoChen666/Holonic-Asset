package imageprocessor

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
)

func TestConcatRejectsInputCountBeforeDecoding(t *testing.T) {
	t.Parallel()

	inputs := make([]ImageInput, limits.MaxConcatInputs+1)
	_, err := New().Concat(inputs, ConcatHorizontal, 0)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Concat error = %v, want ErrResourceLimit", err)
	}
}

func TestConcatRejectsHugeGapWithoutPanic(t *testing.T) {
	t.Parallel()

	input := encodedTestImage(t, 1, 1)
	_, err := New().Concat(
		[]ImageInput{input, input},
		ConcatHorizontal,
		int(^uint(0)>>1),
	)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Concat error = %v, want ErrResourceLimit", err)
	}
}

func TestResizeRejectsHugeOutputBeforeDecoding(t *testing.T) {
	t.Parallel()

	_, err := New().Resize(
		ImageInput{},
		int(^uint(0)>>1),
		2,
		ResizePixelArt,
	)
	if !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("Resize error = %v, want ErrResourceLimit", err)
	}
}

func TestLimitedBufferRejectsEncodedOutputOverLimit(t *testing.T) {
	t.Parallel()

	buffer := limitedBuffer{maximum: 4}
	if _, err := buffer.Write([]byte{1, 2, 3, 4}); err != nil {
		t.Fatalf("limitedBuffer rejected content at the limit: %v", err)
	}
	if _, err := buffer.Write([]byte{5}); !errors.Is(err, limits.ErrResourceLimit) {
		t.Fatalf("limitedBuffer error = %v, want ErrResourceLimit", err)
	}
}

func encodedTestImage(t *testing.T, width int, height int) ImageInput {
	t.Helper()

	var content bytes.Buffer
	if err := png.Encode(&content, image.NewNRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatalf("encode test image: %v", err)
	}
	return ImageInput{Base64: base64.StdEncoding.EncodeToString(content.Bytes())}
}
