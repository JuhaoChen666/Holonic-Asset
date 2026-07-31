// Package limits centralizes resource limits and overflow-safe arithmetic for
// image processor operations.
package limits

import (
	"errors"
	"fmt"
)

const (
	// MaxImageBytes limits one compressed input image.
	MaxImageBytes = 20 << 20
	// MaxImagePixels limits one decoded input image.
	MaxImagePixels = 40_000_000
	// MaxOutputPixels limits one processed output canvas.
	MaxOutputPixels = 40_000_000
	// MaxWorkingPixels limits simultaneously resident decoded source and output pixels.
	MaxWorkingPixels = MaxImagePixels + MaxOutputPixels
	// MaxOutputBytes limits the encoded PNG held in memory.
	MaxOutputBytes = 20 << 20
	// MaxConcatInputs limits fan-in work and per-input bookkeeping.
	MaxConcatInputs = 256
	// MaxConcatEncodedBytes limits compressed inputs retained during Concat preflight.
	MaxConcatEncodedBytes = 64 << 20
)

var ErrResourceLimit = errors.New("image resource limit exceeded")

// PixelCount validates dimensions and returns width*height without overflowing int.
func PixelCount(label string, width int, height int, maximum int) (int, error) {
	if width <= 0 || height <= 0 {
		return 0, fmt.Errorf("%s dimensions must be positive", label)
	}
	if width > maximum/height {
		return 0, fmt.Errorf(
			"%w: %s exceeds the %d-pixel limit",
			ErrResourceLimit,
			label,
			maximum,
		)
	}
	return width * height, nil
}

// CheckedAdd returns left+right or an error when the result would overflow int.
func CheckedAdd(label string, left int, right int) (int, error) {
	if left < 0 || right < 0 {
		return 0, fmt.Errorf("%s values cannot be negative", label)
	}
	maxInt := int(^uint(0) >> 1)
	if left > maxInt-right {
		return 0, fmt.Errorf("%w: %s addition overflows int", ErrResourceLimit, label)
	}
	return left + right, nil
}

// CheckedMultiply returns left*right or an error when the result would overflow int.
func CheckedMultiply(label string, left int, right int) (int, error) {
	if left < 0 || right < 0 {
		return 0, fmt.Errorf("%s values cannot be negative", label)
	}
	if left != 0 && right > int(^uint(0)>>1)/left {
		return 0, fmt.Errorf("%w: %s multiplication overflows int", ErrResourceLimit, label)
	}
	return left * right, nil
}

// CheckMaximum rejects a non-negative value above its configured maximum.
func CheckMaximum(label string, value int, maximum int) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", label)
	}
	if value > maximum {
		return fmt.Errorf(
			"%w: %s exceeds the limit of %d",
			ErrResourceLimit,
			label,
			maximum,
		)
	}
	return nil
}
