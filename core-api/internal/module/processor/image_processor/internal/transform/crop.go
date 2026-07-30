package transform

import (
	"fmt"
	"image"
	"image/draw"
)

func Crop(source *image.NRGBA, rectangle image.Rectangle) (*image.NRGBA, error) {
	if rectangle.Empty() {
		return nil, fmt.Errorf("crop rectangle must have positive dimensions")
	}
	if !rectangle.In(source.Bounds()) {
		return nil, fmt.Errorf("crop rectangle must be inside the image bounds")
	}
	result := image.NewNRGBA(image.Rect(0, 0, rectangle.Dx(), rectangle.Dy()))
	draw.Draw(result, result.Bounds(), source, rectangle.Min, draw.Src)
	return result, nil
}
