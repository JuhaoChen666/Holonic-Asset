package imageprocessor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/layout"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/limits"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/transform"
)

// ImageOutput contains the processed image as PNG-encoded bytes.
type ImageOutput struct {
	Content []byte
	Width   int
	Height  int
	// OffsetX and OffsetY identify the top-left position of a trimmed result
	// in its source canvas. Operations that do not trim leave both values at 0.
	OffsetX int
	OffsetY int
	// Placement identifies where source content was placed on a generated
	// canvas. It is populated by layout-aware operations such as ResizeContain.
	Placement *Rect
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
	ResizeAnchorTopLeft
	ResizeAnchorTopCenter
	ResizeAnchorTopRight
	ResizeAnchorCenterLeft
	ResizeAnchorCenterRight
	ResizeAnchorBottomLeft
	ResizeAnchorBottomRight
	ResizeAnchorCustom
)

type Insets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

type ResizeContainOptions struct {
	Width        int
	Height       int
	Filter       ResizeFilter
	Anchor       ResizeAnchor
	Padding      Insets
	AllowUpscale bool
	// AnchorX and AnchorY are normalized positions from 0 through 1 and are
	// used only when Anchor is ResizeAnchorCustom.
	AnchorX float64
	AnchorY float64
}

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
	// Background may be omitted to infer the dominant matte color from the
	// complete image border.
	Background *RGB
	// Tolerance is applied as provided; zero means exact color matching.
	Tolerance uint8
	// Feather is applied as provided; zero disables edge feathering.
	Feather uint8
	// SpillSuppression removes the sampled matte contribution from partially
	// transparent edge pixels. Values are percentages from 0 through 100.
	SpillSuppression uint8
	// RemoveEnclosed also removes matching matte regions that are not connected
	// to the image border. Enable it only for controlled matte sources whose
	// subject is guaranteed not to use the key color.
	RemoveEnclosed bool
	// ChromaKey derives soft alpha from a strongly dominant matte channel. It
	// is intended for controlled green, blue, or red chroma sources and falls
	// back to distance matching when the sampled matte has no dominant channel.
	// In this mode, Tolerance and Feather are the transparent and opaque alpha
	// dead zones respectively.
	ChromaKey bool
}

type ConcatDirection uint8

const (
	ConcatHorizontal ConcatDirection = iota
	ConcatVertical
)

// Processor is a collection of independent image tools operating on encoded images.
type Processor struct{}

type preparedImage struct {
	content []byte
	config  image.Config
	format  string
}

// New returns a Processor ready to apply image operations on encoded inputs.
func New() *Processor {
	return &Processor{}
}

func (p *Processor) Crop(input []byte, rectangle Rect) (ImageOutput, error) {
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
	input []byte,
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
	input []byte,
	width int,
	height int,
	filter ResizeFilter,
	anchor ResizeAnchor,
) (ImageOutput, error) {
	return p.ResizeContainWithOptions(input, ResizeContainOptions{
		Width:        width,
		Height:       height,
		Filter:       filter,
		Anchor:       anchor,
		AllowUpscale: true,
	})
}

// ResizeContainWithOptions preserves source aspect ratio and places it inside
// a padded transparent canvas. It reports the exact placement rectangle.
func (p *Processor) ResizeContainWithOptions(
	input []byte,
	options ResizeContainOptions,
) (ImageOutput, error) {
	if _, err := limits.PixelCount(
		"resize output",
		options.Width,
		options.Height,
		limits.MaxOutputPixels,
	); err != nil {
		return ImageOutput{}, err
	}
	if err := validateInsets(options.Width, options.Height, options.Padding); err != nil {
		return ImageOutput{}, err
	}
	internalFilter, err := toTransformResizeFilter(options.Filter)
	if err != nil {
		return ImageOutput{}, err
	}
	anchorX, anchorY, err := resizeAnchorPosition(options)
	if err != nil {
		return ImageOutput{}, err
	}
	source, err := p.load(input)
	if err != nil {
		return ImageOutput{}, err
	}
	result, placement, err := transform.ResizeContain(
		source,
		options.Width,
		options.Height,
		internalFilter,
		transform.Insets{
			Top:    options.Padding.Top,
			Right:  options.Padding.Right,
			Bottom: options.Padding.Bottom,
			Left:   options.Padding.Left,
		},
		anchorX,
		anchorY,
		options.AllowUpscale,
	)
	if err != nil {
		return ImageOutput{}, err
	}
	output, err := encodeOutput(result)
	if err != nil {
		return ImageOutput{}, err
	}
	output.Placement = &Rect{
		X:      placement.Min.X,
		Y:      placement.Min.Y,
		Width:  placement.Dx(),
		Height: placement.Dy(),
	}
	return output, nil
}

func (p *Processor) AdjustRGBA(
	input []byte,
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
	input []byte,
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
	input []byte,
	options RemoveBackgroundOptions,
) (ImageOutput, error) {
	if options.SpillSuppression > 100 {
		return ImageOutput{}, fmt.Errorf(
			"spill suppression must be between 0 and 100",
		)
	}
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
		options.SpillSuppression,
		options.RemoveEnclosed,
		options.ChromaKey,
	))
}

func (p *Processor) TrimTransparent(
	input []byte,
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
	inputs [][]byte,
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

func resizeAnchorPosition(options ResizeContainOptions) (float64, float64, error) {
	switch options.Anchor {
	case ResizeAnchorCenter:
		return 0.5, 0.5, nil
	case ResizeAnchorBottomCenter:
		return 0.5, 1, nil
	case ResizeAnchorTopLeft:
		return 0, 0, nil
	case ResizeAnchorTopCenter:
		return 0.5, 0, nil
	case ResizeAnchorTopRight:
		return 1, 0, nil
	case ResizeAnchorCenterLeft:
		return 0, 0.5, nil
	case ResizeAnchorCenterRight:
		return 1, 0.5, nil
	case ResizeAnchorBottomLeft:
		return 0, 1, nil
	case ResizeAnchorBottomRight:
		return 1, 1, nil
	case ResizeAnchorCustom:
		if math.IsNaN(options.AnchorX) || math.IsNaN(options.AnchorY) ||
			math.IsInf(options.AnchorX, 0) || math.IsInf(options.AnchorY, 0) ||
			options.AnchorX < 0 || options.AnchorX > 1 ||
			options.AnchorY < 0 || options.AnchorY > 1 {
			return 0, 0, fmt.Errorf(
				"custom resize anchor coordinates must be finite numbers between 0 and 1",
			)
		}
		return options.AnchorX, options.AnchorY, nil
	default:
		return 0, 0, fmt.Errorf("unsupported resize anchor: %d", options.Anchor)
	}
}

func validateInsets(width int, height int, padding Insets) error {
	if padding.Top < 0 || padding.Right < 0 ||
		padding.Bottom < 0 || padding.Left < 0 {
		return fmt.Errorf("resize padding cannot be negative")
	}
	if padding.Left > width ||
		padding.Right > width-padding.Left ||
		padding.Top > height ||
		padding.Bottom > height-padding.Top ||
		width-padding.Left-padding.Right <= 0 ||
		height-padding.Top-padding.Bottom <= 0 {
		return fmt.Errorf("resize padding must leave a positive content area")
	}
	return nil
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

func (p *Processor) load(input []byte) (*image.NRGBA, error) {
	value, err := prepare(input)
	if err != nil {
		return nil, err
	}
	return decodePrepared(value)
}

func prepare(content []byte) (preparedImage, error) {
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

	config, format, err := image.DecodeConfig(bytes.NewReader(content))
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
	return preparedImage{content: content, config: config, format: format}, nil
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
		Content: append([]byte(nil), content.Bytes()...),
		Width:   value.Bounds().Dx(),
		Height:  value.Bounds().Dy(),
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
