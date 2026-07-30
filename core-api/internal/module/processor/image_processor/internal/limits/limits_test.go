package limits

import (
	"errors"
	"testing"
)

func TestPixelCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		height  int
		want    int
		wantErr bool
	}{
		{name: "valid", width: 32, height: 64, want: 2048},
		{name: "zero width", width: 0, height: 64, wantErr: true},
		{name: "negative height", width: 32, height: -1, wantErr: true},
		{name: "over pixel limit", width: MaxImagePixels, height: 2, wantErr: true},
		{name: "int overflow", width: int(^uint(0) >> 1), height: 2, wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := PixelCount("test image", test.width, test.height, MaxImagePixels)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("PixelCount returned an error: %v", err)
			}
			if got != test.want {
				t.Fatalf("PixelCount = %d, want %d", got, test.want)
			}
		})
	}
}

func TestCheckedArithmeticRejectsOverflow(t *testing.T) {
	t.Parallel()

	maxInt := int(^uint(0) >> 1)
	if _, err := CheckedAdd("sum", maxInt, 1); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("CheckedAdd error = %v, want ErrResourceLimit", err)
	}
	if _, err := CheckedMultiply("product", maxInt, 2); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("CheckedMultiply error = %v, want ErrResourceLimit", err)
	}
}

func TestCheckMaximum(t *testing.T) {
	t.Parallel()

	if err := CheckMaximum("count", MaxConcatInputs, MaxConcatInputs); err != nil {
		t.Fatalf("CheckMaximum rejected the configured maximum: %v", err)
	}
	if err := CheckMaximum("count", MaxConcatInputs+1, MaxConcatInputs); !errors.Is(err, ErrResourceLimit) {
		t.Fatalf("CheckMaximum error = %v, want ErrResourceLimit", err)
	}
}
