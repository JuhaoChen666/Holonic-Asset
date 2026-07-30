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
	"math"
	"strings"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/layout"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/transform"
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
	// OffsetX and OffsetY identify the top-left position of a trimmed result
	// in its source canvas. Operations that do not trim leave both values at 0.
	OffsetX int
	OffsetY int
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

type ResizeAnchor uint8

const (
	ResizeAnchorCenter ResizeAnchor = iota
	ResizeAnchorBottomCenter
)

type RGBAAdjustment struct {
	Red   int
	Green int
	Blue  int
	Alpha int
	// PreserveTransparent prevents an alpha adjustment from making pixels
	// whose original alpha is zero visible again.
	PreserveTransparent bool
}

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

type RemoveBackgroundOptions struct {
	// Background may be omitted to infer the color from the four corners.
	Background *RGB
	// Tolerance is applied as provided; zero means exact color matching.
	Tolerance uint8
	// Feather is applied as provided; zero disables edge feathering.
	Feather uint8
}

type ConcatDirection uint8

const (
	ConcatHorizontal ConcatDirection = iota
	ConcatVertical
)

// Processor is a collection of independent image tools operating on Base64-encoded images.
type Processor struct{}

type preparedImage struct {
	content []byte
	config  image.Config
}

// New returns a Processor ready to apply image operations on Base64-encoded inputs.
func New() *Processor {
	return &Processor{}
}

func (p *Processor) Crop(input ImageInput, rectangle Rect) (ImageOutput, error) {
	if rectangle.Width <= 0 || rectangle.Height <= 0 {
		return ImageOutput{}, fmt.Errorf("crop dimensions must be positive")
	}
	if rectangle.X < 0 || rectangle.Y < 0 {
		return ImageOutput{}, fmt.Errorf("crop coordinates cannot be negative")
	}
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	sourceWidth := source.Bounds().Dx()
	sourceHeight := source.Bounds().Dy()
	if rectangle.X > sourceWidth ||
		rectangle.Y > sourceHeight ||
		rectangle.Width > sourceWidth-rectangle.X ||
		rectangle.Height > sourceHeight-rectangle.Y {
		return ImageOutput{}, fmt.Errorf("crop rectangle must be inside the image bounds")
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
	if _, err := limits.PixelCount(
		"resize output",
		width,
		height,
		limits.MaxOutputPixels,
	); err != nil {
		return ImageOutput{}, err
	}
	internalFilter, err := toTransformResizeFilter(filter)
	if err != nil {
		return ImageOutput{}, err
	}
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, err := transform.Resize(source, width, height, internalFilter)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

// ResizeContain preserves the source aspect ratio, places the scaled image on
// a transparent canvas of exactly width by height, and never stretches it.
func (p *Processor) ResizeContain(
	input ImageInput,
	width int,
	height int,
	filter ResizeFilter,
	anchor ResizeAnchor,
) (ImageOutput, error) {
	if _, err := limits.PixelCount(
		"resize output",
		width,
		height,
		limits.MaxOutputPixels,
	); err != nil {
		return ImageOutput{}, err
	}
	internalFilter, err := toTransformResizeFilter(filter)
	if err != nil {
		return ImageOutput{}, err
	}
	internalAnchor, err := toTransformResizeAnchor(anchor)
	if err != nil {
		return ImageOutput{}, err
	}
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, err := transform.ResizeContain(
		source,
		width,
		height,
		internalFilter,
		internalAnchor,
	)
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
		adjustment.PreserveTransparent,
	))
}

func (p *Processor) SetOpacity(
	input ImageInput,
	opacity float64,
) (ImageOutput, error) {
	if math.IsNaN(opacity) ||
		math.IsInf(opacity, 0) ||
		opacity < 0 ||
		opacity > 1 {
		return ImageOutput{}, fmt.Errorf("opacity must be a finite number between 0 and 1")
	}
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
	var background *color.NRGBA
	if options.Background != nil {
		background = &color.NRGBA{
			R: options.Background.Red,
			G: options.Background.Green,
			B: options.Background.Blue,
			A: 255,
		}
	}
	return encodeOutput(transform.RemoveBackground(
		source,
		background,
		options.Tolerance,
		options.Feather,
	))
}

func (p *Processor) TrimTransparent(
	input ImageInput,
	alphaThreshold uint8,
) (ImageOutput, error) {
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, offset, err := transform.TrimTransparent(source, alphaThreshold)
	if err != nil {
		return ImageOutput{}, err
	}
	output, err := encodeOutput(result)
	if err != nil {
		return ImageOutput{}, err
	}
	output.OffsetX = offset.X
	output.OffsetY = offset.Y
	return output, nil
}

func (p *Processor) Concat(
	inputs []ImageInput,
	direction ConcatDirection,
	gap int,
) (ImageOutput, error) {
	internalDirection, err := toLayoutDirection(direction)
	if err != nil {
		return ImageOutput{}, err
	}
	if err := limits.CheckMaximum(
		"concat input count",
		len(inputs),
		limits.MaxConcatInputs,
	); err != nil {
		return ImageOutput{}, err
	}

	prepared := make([]preparedImage, 0, len(inputs))
	sizes := make([]image.Point, 0, len(inputs))
	totalEncodedBytes := 0
	for index, input := range inputs {
		value, err := prepare(input)
		if err != nil {
			return ImageOutput{}, fmt.Errorf("prepare image %d: %w", index, err)
		}
		totalEncodedBytes, err = limits.CheckedAdd(
			"concat encoded bytes",
			totalEncodedBytes,
			len(value.content),
		)
		if err != nil {
			return ImageOutput{}, err
		}
		if err := limits.CheckMaximum(
			"concat encoded bytes",
			totalEncodedBytes,
			limits.MaxConcatEncodedBytes,
		); err != nil {
			return ImageOutput{}, err
		}
		prepared = append(prepared, value)
		sizes = append(sizes, image.Pt(value.config.Width, value.config.Height))
	}

	if _, _, _, err := layout.Dimensions(sizes, internalDirection, gap); err != nil {
		return ImageOutput{}, err
	}

	sources := make([]*image.NRGBA, 0, len(prepared))
	for index, value := range prepared {
		source, err := decodePrepared(value)
		if err != nil {
			return ImageOutput{}, fmt.Errorf("load image %d: %w", index, err)
		}
		sources = append(sources, source)
	}
	result, err := layout.Concat(sources, internalDirection, gap)
	if err != nil {
		return ImageOutput{}, err
	}
	return encodeOutput(result)
}

func toTransformResizeFilter(filter ResizeFilter) (transform.ResizeFilter, error) {
	switch filter {
	case ResizeSmooth:
		return transform.Smooth, nil
	case ResizePixelArt:
		return transform.PixelArt, nil
	default:
		return 0, fmt.Errorf("unsupported resize filter: %d", filter)
	}
}

func toTransformResizeAnchor(anchor ResizeAnchor) (transform.ResizeAnchor, error) {
	switch anchor {
	case ResizeAnchorCenter:
		return transform.AnchorCenter, nil
	case ResizeAnchorBottomCenter:
		return transform.AnchorBottomCenter, nil
	default:
		return 0, fmt.Errorf("unsupported resize anchor: %d", anchor)
	}
}

func toLayoutDirection(direction ConcatDirection) (layout.Direction, error) {
	switch direction {
	case ConcatHorizontal:
		return layout.Horizontal, nil
	case ConcatVertical:
		return layout.Vertical, nil
	default:
		return 0, fmt.Errorf("unsupported concat direction: %d", direction)
	}
}

func (p *Processor) load(input ImageInput) (*image.NRGBA, error) {
	value, err := prepare(input)
	if err != nil {
		return nil, err
	}
	return decodePrepared(value)
}

func prepare(input ImageInput) (preparedImage, error) {
	if strings.TrimSpace(input.Base64) == "" {
		return preparedImage{}, fmt.Errorf("base64 image content is required")
	}
	content, err := decodeBase64(input.Base64)
	if err != nil {
		return preparedImage{}, err
	}
	if len(content) == 0 {
		return preparedImage{}, fmt.Errorf("image content is empty")
	}
	if err := limits.CheckMaximum(
		"image bytes",
		len(content),
		limits.MaxImageBytes,
	); err != nil {
		return preparedImage{}, err
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return preparedImage{}, fmt.Errorf("decode image config: %w", err)
	}
	if _, err := limits.PixelCount(
		"image",
		config.Width,
		config.Height,
		limits.MaxImagePixels,
	); err != nil {
		return preparedImage{}, err
	}
	return preparedImage{content: content, config: config}, nil
}

func decodePrepared(value preparedImage) (*image.NRGBA, error) {
	decoded, _, err := image.Decode(bytes.NewReader(value.content))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	if _, err := limits.PixelCount(
		"decoded image",
		decoded.Bounds().Dx(),
		decoded.Bounds().Dy(),
		limits.MaxImagePixels,
	); err != nil {
		return nil, err
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
	if err := limits.CheckMaximum(
		"base64 image bytes",
		len(value),
		limits.MaxBase64Bytes,
	); err != nil {
		return nil, err
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
	if value == nil {
		return ImageOutput{}, fmt.Errorf("image output is required")
	}
	if _, err := limits.PixelCount(
		"image output",
		value.Bounds().Dx(),
		value.Bounds().Dy(),
		limits.MaxOutputPixels,
	); err != nil {
		return ImageOutput{}, err
	}
	content := limitedBuffer{maximum: limits.MaxOutputBytes}
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

type limitedBuffer struct {
	bytes.Buffer
	maximum int
}

func (buffer *limitedBuffer) Write(value []byte) (int, error) {
	if len(value) > buffer.maximum-buffer.Len() {
		return 0, fmt.Errorf(
			"%w: encoded PNG exceeds the %d-byte limit",
			limits.ErrResourceLimit,
			buffer.maximum,
		)
	}
	return buffer.Buffer.Write(value)
}

func toNRGBA(source image.Image) *image.NRGBA {
	bounds := source.Bounds()
	result := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)
	return result
}
