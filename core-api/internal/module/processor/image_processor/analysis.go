package imageprocessor

import (
	"fmt"
	"image"
	"sort"
	"strings"
)

const (
	defaultPaletteSize    = 16
	maxPaletteSize        = 256
	defaultComponents     = 64
	maxReportedComponents = 1024
)

type PixelConnectivity uint8

const (
	ConnectivityEight PixelConnectivity = iota
	ConnectivityFour
)

type AnalyzeImageOptions struct {
	AlphaThreshold uint8
	PaletteSize    int
	Connectivity   PixelConnectivity
	MaxComponents  int
}

type ComponentInfo struct {
	Area   int
	Bounds Rect
}

type PaletteColor struct {
	RGB   RGB
	Count int
	Ratio float64
}

type ImageAnalysis struct {
	Width                       int
	Height                      int
	Format                      string
	AlphaMin                    uint8
	AlphaMax                    uint8
	TransparentPixels           int
	PartialAlphaPixels          int
	OpaquePixels                int
	TransparentRatio            float64
	PartialAlphaRatio           float64
	OpaqueRatio                 float64
	ContentPixels               int
	ContentBounds               *Rect
	TransparentRGBResiduePixels int
	TouchesEdge                 bool
	EdgeContentPixels           int
	ComponentCount              int
	LargestComponentArea        int
	LargestComponentRatio       float64
	Components                  []ComponentInfo
	Palette                     []PaletteColor
}

func (p *Processor) AnalyzeImage(
	input []byte,
	options AnalyzeImageOptions,
) (ImageAnalysis, error) {
	normalized, err := normalizeAnalyzeOptions(options)
	if err != nil {
		return ImageAnalysis{}, err
	}
	prepared, err := prepare(input)
	if err != nil {
		return ImageAnalysis{}, err
	}
	source, err := decodePrepared(prepared)
	if err != nil {
		return ImageAnalysis{}, err
	}
	return analyzeImage(source, prepared.format, normalized), nil
}

func analyzeImage(
	source *image.NRGBA,
	format string,
	options AnalyzeImageOptions,
) ImageAnalysis {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	totalPixels := width * height
	analysis := ImageAnalysis{
		Width:    width,
		Height:   height,
		Format:   strings.ToLower(format),
		AlphaMin: 255,
	}
	type paletteBucket struct {
		red   int
		green int
		blue  int
		count int
	}
	palette := make(map[uint16]paletteBucket)
	left, top, right, bottom := width, height, 0, 0
	hasContent := false
	for y := range height {
		for x := range width {
			pixel := source.NRGBAAt(x, y)
			analysis.AlphaMin = min(analysis.AlphaMin, pixel.A)
			analysis.AlphaMax = max(analysis.AlphaMax, pixel.A)
			switch pixel.A {
			case 0:
				analysis.TransparentPixels++
				if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
					analysis.TransparentRGBResiduePixels++
				}
			case 255:
				analysis.OpaquePixels++
			default:
				analysis.PartialAlphaPixels++
			}
			if pixel.A <= options.AlphaThreshold {
				continue
			}
			analysis.ContentPixels++
			hasContent = true
			left = min(left, x)
			top = min(top, y)
			right = max(right, x+1)
			bottom = max(bottom, y+1)
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				analysis.EdgeContentPixels++
			}
			key := uint16(pixel.R>>3)<<10 |
				uint16(pixel.G>>3)<<5 |
				uint16(pixel.B>>3)
			bucket := palette[key]
			bucket.red += int(pixel.R)
			bucket.green += int(pixel.G)
			bucket.blue += int(pixel.B)
			bucket.count++
			palette[key] = bucket
		}
	}
	analysis.TransparentRatio = imageRatio(
		analysis.TransparentPixels,
		totalPixels,
	)
	analysis.PartialAlphaRatio = imageRatio(
		analysis.PartialAlphaPixels,
		totalPixels,
	)
	analysis.OpaqueRatio = imageRatio(analysis.OpaquePixels, totalPixels)
	analysis.TouchesEdge = analysis.EdgeContentPixels > 0
	if hasContent {
		analysis.ContentBounds = &Rect{
			X: left, Y: top, Width: right - left, Height: bottom - top,
		}
	}

	components := make([]ComponentInfo, 0, options.MaxComponents)
	walkAlphaComponents(
		source,
		options.AlphaThreshold,
		options.Connectivity,
		func(component ComponentInfo, _ []componentSpan) {
			analysis.ComponentCount++
			analysis.LargestComponentArea = max(
				analysis.LargestComponentArea,
				component.Area,
			)
			if len(components) < options.MaxComponents {
				components = append(components, component)
				return
			}
			worst := 0
			for index := 1; index < len(components); index++ {
				if componentComesBefore(components[worst], components[index]) {
					worst = index
				}
			}
			if componentComesBefore(component, components[worst]) {
				components[worst] = component
			}
		},
	)
	sort.Slice(components, func(left int, right int) bool {
		return componentComesBefore(components[left], components[right])
	})
	if analysis.ComponentCount > 0 {
		analysis.LargestComponentRatio = imageRatio(
			analysis.LargestComponentArea,
			analysis.ContentPixels,
		)
	}
	analysis.Components = append(
		[]ComponentInfo(nil),
		components...,
	)

	analysis.Palette = make([]PaletteColor, 0, len(palette))
	for _, bucket := range palette {
		analysis.Palette = append(analysis.Palette, PaletteColor{
			RGB: RGB{
				Red:   averagePaletteChannel(bucket.red, bucket.count),
				Green: averagePaletteChannel(bucket.green, bucket.count),
				Blue:  averagePaletteChannel(bucket.blue, bucket.count),
			},
			Count: bucket.count,
			Ratio: imageRatio(bucket.count, analysis.ContentPixels),
		})
	}
	sort.Slice(analysis.Palette, func(left int, right int) bool {
		if analysis.Palette[left].Count != analysis.Palette[right].Count {
			return analysis.Palette[left].Count > analysis.Palette[right].Count
		}
		leftRGB := analysis.Palette[left].RGB
		rightRGB := analysis.Palette[right].RGB
		if leftRGB.Red != rightRGB.Red {
			return leftRGB.Red < rightRGB.Red
		}
		if leftRGB.Green != rightRGB.Green {
			return leftRGB.Green < rightRGB.Green
		}
		return leftRGB.Blue < rightRGB.Blue
	})
	if len(analysis.Palette) > options.PaletteSize {
		analysis.Palette = analysis.Palette[:options.PaletteSize]
	}
	return analysis
}

func averagePaletteChannel(total int, count int) uint8 {
	if count <= 0 {
		return 0
	}
	average := total / count
	if average <= 0 {
		return 0
	}
	if average >= 255 {
		return 255
	}
	return uint8(average)
}

func normalizeAnalyzeOptions(
	options AnalyzeImageOptions,
) (AnalyzeImageOptions, error) {
	if options.PaletteSize < 0 || options.PaletteSize > maxPaletteSize {
		return AnalyzeImageOptions{}, fmt.Errorf(
			"palette size must be between %d and %d: %d",
			0,
			maxPaletteSize,
			options.PaletteSize,
		)
	}
	if options.MaxComponents < 0 ||
		options.MaxComponents > maxReportedComponents {
		return AnalyzeImageOptions{}, fmt.Errorf(
			"maximum reported components must be between %d and %d: %d",
			0,
			maxReportedComponents,
			options.MaxComponents,
		)
	}
	if !validConnectivity(options.Connectivity) {
		return AnalyzeImageOptions{}, fmt.Errorf(
			"unsupported pixel connectivity: %d",
			options.Connectivity,
		)
	}
	if options.PaletteSize == 0 {
		options.PaletteSize = defaultPaletteSize
	}
	if options.MaxComponents == 0 {
		options.MaxComponents = defaultComponents
	}
	return options, nil
}

type componentSpan struct {
	y     int
	left  int
	right int
}

func walkAlphaComponents(
	source *image.NRGBA,
	alphaThreshold uint8,
	connectivity PixelConnectivity,
	visit func(ComponentInfo, []componentSpan),
) {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	visited := make([]bool, width*height)
	visible := func(x int, y int) bool {
		return source.NRGBAAt(x, y).A > alphaThreshold
	}
	expansion := 1
	if connectivity == ConnectivityFour {
		expansion = 0
	}
	stack := make([]image.Point, 0)
	spans := make([]componentSpan, 0)

	for startY := range height {
		for startX := range width {
			startIndex := startY*width + startX
			if visited[startIndex] || !visible(startX, startY) {
				continue
			}
			stack = append(stack[:0], image.Pt(startX, startY))
			spans = spans[:0]
			area := 0
			left, top, right, bottom := width, height, 0, 0
			for len(stack) > 0 {
				point := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				index := point.Y*width + point.X
				if visited[index] || !visible(point.X, point.Y) {
					continue
				}

				runLeft := point.X
				for runLeft > 0 {
					next := point.Y*width + runLeft - 1
					if visited[next] || !visible(runLeft-1, point.Y) {
						break
					}
					runLeft--
				}
				runRight := point.X
				for runRight+1 < width {
					next := point.Y*width + runRight + 1
					if visited[next] || !visible(runRight+1, point.Y) {
						break
					}
					runRight++
				}
				for x := runLeft; x <= runRight; x++ {
					visited[point.Y*width+x] = true
				}
				spans = append(spans, componentSpan{
					y: point.Y, left: runLeft, right: runRight,
				})
				runArea := runRight - runLeft + 1
				area += runArea
				left = min(left, runLeft)
				top = min(top, point.Y)
				right = max(right, runRight+1)
				bottom = max(bottom, point.Y+1)

				for _, adjacentY := range [...]int{point.Y - 1, point.Y + 1} {
					if adjacentY < 0 || adjacentY >= height {
						continue
					}
					scanLeft := max(0, runLeft-expansion)
					scanRight := min(width-1, runRight+expansion)
					for x := scanLeft; x <= scanRight; {
						next := adjacentY*width + x
						if visited[next] || !visible(x, adjacentY) {
							x++
							continue
						}
						stack = append(stack, image.Pt(x, adjacentY))
						x++
						for x <= scanRight &&
							!visited[adjacentY*width+x] &&
							visible(x, adjacentY) {
							x++
						}
					}
				}
			}
			visit(ComponentInfo{
				Area: area,
				Bounds: Rect{
					X: left, Y: top, Width: right - left, Height: bottom - top,
				},
			}, spans)
		}
	}
}

func validConnectivity(connectivity PixelConnectivity) bool {
	return connectivity == ConnectivityEight ||
		connectivity == ConnectivityFour
}

func componentComesBefore(left ComponentInfo, right ComponentInfo) bool {
	if left.Area != right.Area {
		return left.Area > right.Area
	}
	if left.Bounds.Y != right.Bounds.Y {
		return left.Bounds.Y < right.Bounds.Y
	}
	return left.Bounds.X < right.Bounds.X
}

func imageRatio(numerator int, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
