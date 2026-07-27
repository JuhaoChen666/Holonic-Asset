package imageprocessor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"strings"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/layout"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/transform"
)

const (
	maxImageBytes  = 20 << 20
	maxImagePixels = 40_000_000
)

// ImageInput identifies an image by its Base64-encoded content.
type ImageInput struct {
	Base64 string
}

// ImageOutput contains the processed image as Base64-encoded PNG content.
type ImageOutput struct {
	Base64    string
	MediaType string
	Width     int
	Height    int
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type ResizeFilter uint8

const (
	ResizeSmooth ResizeFilter = iota
	ResizePixelArt
)

type RGBAAdjustment struct {
	Red   int
	Green int
	Blue  int
	Alpha int
}

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

type RemoveBackgroundOptions struct {
	// Background may be omitted to infer the color from the four corners.
	Background *RGB
	// A zero tolerance uses the default value of 12.
	Tolerance uint8
	// A zero feather uses the default value of 8.
	Feather uint8
}

type ConcatDirection uint8

const (
	ConcatHorizontal ConcatDirection = iota
	ConcatVertical
)

// Processor is a collection of independent image tools operating on Base64-encoded images.
type Processor struct{}

// New returns a Processor ready to apply image operations on Base64-encoded inputs.
func New() *Processor {
	return &Processor{}
}

func (p *Processor) Crop(input ImageInput, rectangle Rect) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, err := transform.Crop(
		source,
		image.Rect(
			rectangle.X,
			rectangle.Y,
			rectangle.X+rectangle.Width,
			rectangle.Y+rectangle.Height,
		),
	)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func (p *Processor) Resize(
	input ImageInput,
	width int,
	height int,
	filter ResizeFilter,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	internalFilter := transform.Smooth
	if filter == ResizePixelArt {
		internalFilter = transform.PixelArt
	}
	result, err := transform.Resize(source, width, height, internalFilter)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func (p *Processor) AdjustRGBA(
	input ImageInput,
	adjustment RGBAAdjustment,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(transform.AdjustRGBA(
		source,
		adjustment.Red,
		adjustment.Green,
		adjustment.Blue,
		adjustment.Alpha,
	))
}

func (p *Processor) SetOpacity(
	input ImageInput,
	opacity float64,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, err := transform.SetOpacity(source, opacity)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func (p *Processor) RemoveBackground(
	input ImageInput,
	options RemoveBackgroundOptions,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	tolerance := options.Tolerance
	if tolerance == 0 {
		tolerance = 12
	}
	feather := options.Feather
	if feather == 0 {
		feather = 8
	}
	var background *color.NRGBA
	if options.Background != nil {
		background = &color.NRGBA{
			R: options.Background.Red,
			G: options.Background.Green,
			B: options.Background.Blue,
			A: 255,
		}
	}
	return encodeOutput(transform.RemoveBackground(source, background, tolerance, feather))
}

func (p *Processor) TrimTransparent(
	input ImageInput,
	alphaThreshold uint8,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, err := transform.TrimTransparent(source, alphaThreshold)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func (p *Processor) Concat(
	inputs []ImageInput,
	direction ConcatDirection,
	gap int,
) (ImageOutput, error) {
	sources := make([]*image.NRGBA, 0, len(inputs))
	for index, input := range inputs {
		source, err := p.load(input)
		if err != nil {
			return ImageOutput{}, fmt.Errorf("load image %d: %w", index, err)
		}
		sources = append(sources, source)
	}
	internalDirection := layout.Horizontal
	if direction == ConcatVertical {
		internalDirection = layout.Vertical
	}
	result, err := layout.Concat(sources, internalDirection, gap)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func (p *Processor) load(input ImageInput) (*image.NRGBA, error) {
	if strings.TrimSpace(input.Base64) == "" {
		return nil, fmt.Errorf("base64 image content is required")
	}
	content, err := decodeBase64(input.Base64)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("image content is empty")
	}
	if len(content) > maxImageBytes {
		return nil, fmt.Errorf("image exceeds the %d-byte limit", maxImageBytes)
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("decode image config: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 ||
		int64(config.Width)*int64(config.Height) > maxImagePixels {
		return nil, fmt.Errorf("image dimensions exceed the supported limit")
	}
	decoded, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return toNRGBA(decoded), nil
}

func decodeBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "data:") {
		comma := strings.IndexByte(value, ',')
		if comma < 0 || !strings.Contains(value[:comma], ";base64") {
			return nil, fmt.Errorf("invalid image data URI")
		}
		value = value[comma+1:]
	}
	if len(value) > maxImageBytes*2 {
		return nil, fmt.Errorf("base64 image exceeds the supported limit")
	}
	content, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		content, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil, fmt.Errorf("decode image Base64: %w", err)
	}
	return content, nil
}

func encodeOutput(value *image.NRGBA) (ImageOutput, error) {
	var content bytes.Buffer
	if err := png.Encode(&content, value); err != nil {
		return ImageOutput{}, fmt.Errorf("encode PNG: %w", err)
	}
	return ImageOutput{
		Base64:    base64.StdEncoding.EncodeToString(content.Bytes()),
		MediaType: "image/png",
		Width:     value.Bounds().Dx(),
		Height:    value.Bounds().Dy(),
	}, nil
}

func toNRGBA(source image.Image) *image.NRGBA {
	bounds := source.Bounds()
	result := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)
	return result
}
