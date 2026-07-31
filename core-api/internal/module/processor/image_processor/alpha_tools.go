package imageprocessor

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image_processor/internal/transform"
)

type AlphaBoundsResult struct {
	Found         bool
	Bounds        *Rect
	ContentPixels int
	TouchesEdge   bool
}

// AlphaBounds analyzes visible pixels without cropping or mutating the image.
func (p *Processor) AlphaBounds(
	input []byte,
	alphaThreshold uint8,
) (AlphaBoundsResult, error) {
	source, err := p.load(input)
	if err != nil {
		return AlphaBoundsResult{}, err
	}
	bounds := transform.AlphaBounds(source, alphaThreshold)
	result := AlphaBoundsResult{
		Found:         bounds.Found,
		ContentPixels: bounds.ContentPixels,
		TouchesEdge:   bounds.TouchesEdge,
	}
	if bounds.Found {
		result.Bounds = &Rect{
			X:      bounds.Bounds.Min.X,
			Y:      bounds.Bounds.Min.Y,
			Width:  bounds.Bounds.Dx(),
			Height: bounds.Bounds.Dy(),
		}
	}
	return result, nil
}

type SanitizeAlphaOptions struct {
	// TransparentThreshold maps alpha values at or below the threshold to a
	// fully transparent black pixel. Zero still scrubs RGB under alpha zero.
	TransparentThreshold uint8
	// OpaqueThreshold maps alpha values at or above the threshold to 255.
	// Zero disables the near-opaque dead zone.
	OpaqueThreshold uint8
}

type SanitizeAlphaStats struct {
	TransparentRGBCleared int
	PixelsMadeTransparent int
	PixelsMadeOpaque      int
}

type SanitizeAlphaResult struct {
	Image ImageOutput
	SanitizeAlphaStats
}

func (p *Processor) SanitizeAlpha(
	input []byte,
	options SanitizeAlphaOptions,
) (SanitizeAlphaResult, error) {
	if options.OpaqueThreshold > 0 &&
		options.OpaqueThreshold <= options.TransparentThreshold {
		return SanitizeAlphaResult{}, fmt.Errorf(
			"opaque alpha threshold must be greater than transparent threshold",
		)
	}
	source, err := p.load(input)
	if err != nil {
		return SanitizeAlphaResult{}, err
	}
	result, stats := sanitizeAlpha(source, options)
	output, err := encodeOutput(result)
	if err != nil {
		return SanitizeAlphaResult{}, err
	}
	return SanitizeAlphaResult{
		Image:              output,
		SanitizeAlphaStats: stats,
	}, nil
}

type CleanComponentsOptions struct {
	AlphaThreshold uint8
	// MinArea is the minimum connected visible-pixel area to retain.
	MinArea      int
	Connectivity PixelConnectivity
}

type CleanComponentsStats struct {
	RemovedComponents []ComponentInfo
	RemovedCount      int
	RemovedArea       int
}

type CleanComponentsResult struct {
	Image ImageOutput
	CleanComponentsStats
}

func (p *Processor) CleanComponents(
	input []byte,
	options CleanComponentsOptions,
) (CleanComponentsResult, error) {
	if options.MinArea <= 0 {
		return CleanComponentsResult{}, fmt.Errorf(
			"minimum component area must be positive",
		)
	}
	if !validConnectivity(options.Connectivity) {
		return CleanComponentsResult{}, fmt.Errorf(
			"unsupported pixel connectivity: %d",
			options.Connectivity,
		)
	}
	source, err := p.load(input)
	if err != nil {
		return CleanComponentsResult{}, err
	}
	result, cleaned := cleanComponents(source, options)
	output, err := encodeOutput(result)
	if err != nil {
		return CleanComponentsResult{}, err
	}
	return CleanComponentsResult{
		Image:                output,
		CleanComponentsStats: cleaned,
	}, nil
}

func sanitizeAlpha(
	source *image.NRGBA,
	options SanitizeAlphaOptions,
) (*image.NRGBA, SanitizeAlphaStats) {
	result := transform.Clone(source)
	stats := SanitizeAlphaStats{}
	width := result.Bounds().Dx()
	height := result.Bounds().Dy()
	for y := range height {
		for x := range width {
			pixel := result.NRGBAAt(x, y)
			original := pixel
			switch {
			case pixel.A <= options.TransparentThreshold:
				if pixel.A > 0 {
					stats.PixelsMadeTransparent++
				}
				if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
					stats.TransparentRGBCleared++
				}
				pixel = color.NRGBA{}
			case options.OpaqueThreshold > 0 &&
				pixel.A >= options.OpaqueThreshold:
				if pixel.A < 255 {
					stats.PixelsMadeOpaque++
				}
				pixel.A = 255
			}
			if pixel != original {
				result.SetNRGBA(x, y, pixel)
			}
		}
	}
	return result, stats
}

func cleanComponents(
	source *image.NRGBA,
	options CleanComponentsOptions,
) (*image.NRGBA, CleanComponentsStats) {
	result := transform.Clone(source)
	cleaned := CleanComponentsStats{}
	walkAlphaComponents(
		result,
		options.AlphaThreshold,
		options.Connectivity,
		func(component ComponentInfo, spans []componentSpan) {
			if component.Area >= options.MinArea {
				return
			}
			cleaned.RemovedComponents = append(
				cleaned.RemovedComponents,
				component,
			)
			cleaned.RemovedArea += component.Area
			for _, span := range spans {
				for x := span.left; x <= span.right; x++ {
					result.SetNRGBA(x, span.y, color.NRGBA{})
				}
			}
		},
	)
	sort.Slice(cleaned.RemovedComponents, func(left int, right int) bool {
		return componentComesBefore(
			cleaned.RemovedComponents[left],
			cleaned.RemovedComponents[right],
		)
	})
	cleaned.RemovedCount = len(cleaned.RemovedComponents)
	return result, cleaned
}
