package imageprocessor

import (
	"image"
	"image/color"
	"math"
)

type AnalyzeTransparencyOptions struct {
	ExpectedMatte *RGB
}

type TransparencyAnalysis struct {
	Width                       int
	Height                      int
	AlphaMin                    uint8
	AlphaMax                    uint8
	TransparentPixels           int
	PartialAlphaPixels          int
	OpaquePixels                int
	TransparentRatio            float64
	CheckerboardDetected        bool
	TouchesEdge                 bool
	EdgeMargin                  int
	StrayPixelCount             int
	LargestComponentArea        int
	LargestComponentRatio       float64
	MatteResidue                float64
	PartialAlphaMatteResidue    float64
	TransparentRGBResiduePixels int
	ContentBounds               *Rect
}

// AnalyzeTransparency measures transparency-related image properties without
// mutating the image or applying acceptance policy.
func (p *Processor) AnalyzeTransparency(
	input []byte,
	options AnalyzeTransparencyOptions,
) (TransparencyAnalysis, error) {
	source, err := p.load(input)
	if err != nil {
		return TransparencyAnalysis{}, err
	}
	return analyzeTransparency(source, options), nil
}

func analyzeTransparency(
	source *image.NRGBA,
	options AnalyzeTransparencyOptions,
) TransparencyAnalysis {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	total := width * height
	analysis := TransparencyAnalysis{
		Width:      width,
		Height:     height,
		AlphaMin:   255,
		EdgeMargin: -1,
	}

	visible := make([]bool, total)
	visiblePixels := 0
	transparentPixels := 0
	left, top, right, bottom := width, height, 0, 0
	expectedMatte := color.NRGBA{}
	hasExpectedMatte := options.ExpectedMatte != nil
	if hasExpectedMatte {
		expectedMatte = color.NRGBA{
			R: options.ExpectedMatte.Red,
			G: options.ExpectedMatte.Green,
			B: options.ExpectedMatte.Blue,
			A: 255,
		}
	}
	matteResidue := float64(0)
	haloResidue := float64(0)
	partialForHalo := 0

	for y := range height {
		for x := range width {
			pixel := source.NRGBAAt(x, y)
			analysis.AlphaMin = min(analysis.AlphaMin, pixel.A)
			analysis.AlphaMax = max(analysis.AlphaMax, pixel.A)
			switch pixel.A {
			case 0:
				transparentPixels++
				if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
					analysis.TransparentRGBResiduePixels++
				}
			case 255:
				analysis.OpaquePixels++
			default:
				analysis.PartialAlphaPixels++
			}
			if pixel.A == 0 {
				continue
			}

			visible[y*width+x] = true
			visiblePixels++
			left = min(left, x)
			top = min(top, y)
			right = max(right, x+1)
			bottom = max(bottom, y+1)
			if hasExpectedMatte {
				residue := matteCloseness(pixel, expectedMatte) *
					(float64(pixel.A) / 255)
				matteResidue += residue
				if pixel.A < 255 {
					partialForHalo++
					haloResidue += residue
				}
			}
		}
	}

	analysis.TransparentPixels = transparentPixels
	analysis.TransparentRatio = ratio(transparentPixels, total)
	if visiblePixels > 0 {
		analysis.ContentBounds = &Rect{
			X: left, Y: top, Width: right - left, Height: bottom - top,
		}
		analysis.EdgeMargin = min(left, top, width-right, height-bottom)
		analysis.TouchesEdge = analysis.EdgeMargin == 0
	}
	if hasExpectedMatte && visiblePixels > 0 {
		analysis.MatteResidue = matteResidue / float64(visiblePixels)
	}
	if hasExpectedMatte && partialForHalo > 0 {
		analysis.PartialAlphaMatteResidue =
			haloResidue / float64(partialForHalo)
	}

	largest, stray := componentMetrics(visible, width, height)
	analysis.StrayPixelCount = stray
	analysis.LargestComponentArea = largest
	analysis.LargestComponentRatio = ratio(largest, visiblePixels)
	analysis.CheckerboardDetected = detectCheckerboard(source)
	return analysis
}

func componentMetrics(
	visible []bool,
	width int,
	height int,
) (largest int, stray int) {
	visited := make([]bool, len(visible))
	totalVisible := 0
	directions := [...]image.Point{
		{X: -1, Y: -1},
		{X: 0, Y: -1},
		{X: 1, Y: -1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
		{X: -1, Y: 1},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}
	for index, isVisible := range visible {
		if !isVisible {
			continue
		}
		totalVisible++
		if visited[index] {
			continue
		}
		visited[index] = true
		queue := []int{index}
		size := 0
		for head := 0; head < len(queue); head++ {
			current := queue[head]
			size++
			x := current % width
			y := current / width
			for _, direction := range directions {
				nextX := x + direction.X
				nextY := y + direction.Y
				if nextX < 0 || nextX >= width ||
					nextY < 0 || nextY >= height {
					continue
				}
				next := nextY*width + nextX
				if visible[next] && !visited[next] {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
		largest = max(largest, size)
	}
	return largest, totalVisible - largest
}

func detectCheckerboard(source *image.NRGBA) bool {
	if source.Bounds().Dx() < 16 || source.Bounds().Dy() < 16 {
		return false
	}
	for _, blockSize := range [...]int{4, 8, 16} {
		matches := 0
		samples := 0
		var evenColor color.NRGBA
		var oddColor color.NRGBA
		haveEven := false
		haveOdd := false
		for y := blockSize / 2; y < source.Bounds().Dy(); y += blockSize {
			for x := blockSize / 2; x < source.Bounds().Dx(); x += blockSize {
				pixel := source.NRGBAAt(x, y)
				if pixel.A < 250 {
					continue
				}
				even := (x/blockSize+y/blockSize)%2 == 0
				if even && !haveEven {
					evenColor, haveEven = pixel, true
				}
				if !even && !haveOdd {
					oddColor, haveOdd = pixel, true
				}
				if even && colorDistanceNRGBA(pixel, evenColor) <= 8 ||
					!even && colorDistanceNRGBA(pixel, oddColor) <= 8 {
					matches++
				}
				samples++
			}
		}
		if !haveEven || !haveOdd || samples < 16 {
			continue
		}
		colorDifference := colorDistanceNRGBA(evenColor, oddColor)
		if colorDifference >= 12 && colorDifference <= 96 &&
			ratio(matches, samples) >= 0.85 {
			return true
		}
	}
	return false
}

func matteCloseness(pixel color.NRGBA, matte color.NRGBA) float64 {
	distance := float64(colorDistanceNRGBA(pixel, matte))
	return clampUnit(1 - distance/128)
}

func colorDistanceNRGBA(left color.NRGBA, right color.NRGBA) int {
	return max(
		absInt(int(left.R)-int(right.R)),
		absInt(int(left.G)-int(right.G)),
		absInt(int(left.B)-int(right.B)),
	)
}

func ratio(numerator int, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func clampUnit(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
