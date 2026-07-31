package transform

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func RemoveBackground(
	source *image.NRGBA,
	background *color.NRGBA,
	tolerance uint8,
	feather uint8,
	spillSuppression uint8,
	removeEnclosed bool,
	chromaKey bool,
) *image.NRGBA {
	result := Clone(source)
	backgroundColor := borderDominantColor(result)
	if background != nil {
		backgroundColor = *background
	}
	if chromaKey && removeDominantChannelMatte(
		result,
		backgroundColor,
		tolerance,
		feather,
		spillSuppression,
	) {
		scrubTransparentRGB(result)
		return result
	}
	removeConnectedMatte(
		result,
		backgroundColor,
		int(tolerance),
		int(feather),
		spillSuppression,
		removeEnclosed,
	)
	scrubTransparentRGB(result)
	return result
}

func removeDominantChannelMatte(
	result *image.NRGBA,
	background color.NRGBA,
	transparentThreshold uint8,
	opaqueThreshold uint8,
	spillSuppression uint8,
) bool {
	matteChannels := [...]uint8{background.R, background.G, background.B}
	dominantIndex := 0
	for index := 1; index < len(matteChannels); index++ {
		if matteChannels[index] > matteChannels[dominantIndex] {
			dominantIndex = index
		}
	}
	second := uint8(0)
	for index, channel := range matteChannels {
		if index != dominantIndex {
			second = max(second, channel)
		}
	}
	dominance := int(matteChannels[dominantIndex]) - int(second)
	if dominance < 64 {
		return false
	}

	width := result.Bounds().Dx()
	height := result.Bounds().Dy()
	for y := range height {
		for x := range width {
			pixel := result.NRGBAAt(x, y)
			channels := [...]uint8{pixel.R, pixel.G, pixel.B}
			other := uint8(0)
			for index, channel := range channels {
				if index != dominantIndex {
					other = max(other, channel)
				}
			}
			excess := int(channels[dominantIndex]) - int(other)
			if excess <= 0 {
				continue
			}
			alphaRatio := 1 - min(float64(excess)/float64(dominance), 1)
			if alphaRatio*255 <= float64(transparentThreshold) {
				alphaRatio = 0
			} else if (1-alphaRatio)*255 <= float64(opaqueThreshold) {
				alphaRatio = 1
			}
			newAlpha := uint8(math.Round(float64(pixel.A) * alphaRatio))
			pixel = suppressMatteSpill(
				pixel,
				background,
				newAlpha,
				spillSuppression,
			)
			pixel.A = newAlpha
			result.SetNRGBA(x, y, pixel)
		}
	}
	return true
}

func TrimTransparent(
	source *image.NRGBA,
	alphaThreshold uint8,
) (*image.NRGBA, image.Point, error) {
	analysis := AlphaBounds(source, alphaThreshold)
	if !analysis.Found {
		return nil, image.Point{}, fmt.Errorf("image has no pixels above the alpha threshold")
	}
	bounds := analysis.Bounds
	result := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)
	return result, bounds.Min, nil
}

type AlphaBoundsInfo struct {
	Bounds        image.Rectangle
	Found         bool
	ContentPixels int
	TouchesEdge   bool
}

func AlphaBounds(source *image.NRGBA, threshold uint8) AlphaBoundsInfo {
	width, height := source.Bounds().Dx(), source.Bounds().Dy()
	left, top, right, bottom := width, height, 0, 0
	result := AlphaBoundsInfo{}
	for y := range height {
		for x := range width {
			if source.NRGBAAt(x, y).A <= threshold {
				continue
			}
			result.Found = true
			result.ContentPixels++
			left = min(left, x)
			top = min(top, y)
			right = max(right, x+1)
			bottom = max(bottom, y+1)
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				result.TouchesEdge = true
			}
		}
	}
	if result.Found {
		result.Bounds = image.Rect(left, top, right, bottom)
	}
	return result
}

func removeConnectedMatte(
	result *image.NRGBA,
	background color.NRGBA,
	tolerance int,
	feather int,
	spillSuppression uint8,
	removeEnclosed bool,
) {
	width := result.Bounds().Dx()
	height := result.Bounds().Dy()
	maximumDistance := tolerance + feather
	visited := make([]bool, width*height)
	queue := make([]image.Point, 0, 2*(width+height))

	enqueue := func(x int, y int) {
		index := y*width + x
		if visited[index] {
			return
		}
		pixel := result.NRGBAAt(x, y)
		if pixel.A == 0 || colorDistance(pixel, background) <= maximumDistance {
			visited[index] = true
			queue = append(queue, image.Pt(x, y))
		}
	}
	for x := range width {
		enqueue(x, 0)
		if height > 1 {
			enqueue(x, height-1)
		}
	}
	for y := 1; y < height-1; y++ {
		enqueue(0, y)
		if width > 1 {
			enqueue(width-1, y)
		}
	}

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
	applyMatte := func(x int, y int) {
		pixel := result.NRGBAAt(x, y)
		distance := colorDistance(pixel, background)
		switch {
		case pixel.A == 0 || distance <= tolerance:
			pixel = color.NRGBA{}
		case feather > 0 && distance <= maximumDistance:
			ratio := float64(distance-tolerance) / float64(feather)
			newAlpha := uint8(math.Round(float64(pixel.A) * ratio))
			pixel = suppressMatteSpill(
				pixel,
				background,
				newAlpha,
				spillSuppression,
			)
			pixel.A = newAlpha
		}
		result.SetNRGBA(x, y, pixel)
	}
	for len(queue) > 0 {
		point := queue[0]
		queue = queue[1:]
		applyMatte(point.X, point.Y)

		for _, direction := range directions {
			x := point.X + direction.X
			y := point.Y + direction.Y
			if x < 0 || x >= width || y < 0 || y >= height {
				continue
			}
			enqueue(x, y)
		}
	}
	if !removeEnclosed {
		return
	}
	for y := range height {
		for x := range width {
			index := y*width + x
			if visited[index] {
				continue
			}
			pixel := result.NRGBAAt(x, y)
			if pixel.A == 0 ||
				colorDistance(pixel, background) <= maximumDistance {
				applyMatte(x, y)
			}
		}
	}
}

func borderDominantColor(source *image.NRGBA) color.NRGBA {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	type colorBin struct {
		count int
		red   int
		green int
		blue  int
	}
	bins := make(map[uint16]colorBin)
	add := func(x int, y int) {
		pixel := source.NRGBAAt(x, y)
		key := uint16(pixel.R>>4)<<8 |
			uint16(pixel.G>>4)<<4 |
			uint16(pixel.B>>4)
		bin := bins[key]
		bin.count++
		bin.red += int(pixel.R)
		bin.green += int(pixel.G)
		bin.blue += int(pixel.B)
		bins[key] = bin
	}
	for x := range width {
		add(x, 0)
		if height > 1 {
			add(x, height-1)
		}
	}
	for y := 1; y < height-1; y++ {
		add(0, y)
		if width > 1 {
			add(width-1, y)
		}
	}

	var dominant colorBin
	for _, bin := range bins {
		if bin.count > dominant.count {
			dominant = bin
		}
	}
	if dominant.count == 0 {
		return color.NRGBA{A: 255}
	}
	return color.NRGBA{
		R: averageColorChannel(dominant.red, dominant.count),
		G: averageColorChannel(dominant.green, dominant.count),
		B: averageColorChannel(dominant.blue, dominant.count),
		A: 255,
	}
}

func averageColorChannel(total int, count int) uint8 {
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

func suppressMatteSpill(
	pixel color.NRGBA,
	background color.NRGBA,
	alpha uint8,
	strength uint8,
) color.NRGBA {
	if strength == 0 || alpha == 0 || alpha == 255 {
		return pixel
	}
	alphaRatio := float64(alpha) / 255
	strengthRatio := min(float64(strength)/100, 1)
	uncomposite := func(value uint8, matte uint8) uint8 {
		foreground := (float64(value) -
			(1-alphaRatio)*float64(matte)) / alphaRatio
		foreground = max(float64(0), min(float64(255), foreground))
		blended := float64(value) +
			(foreground-float64(value))*strengthRatio
		return uint8(math.Round(blended))
	}
	pixel.R = uncomposite(pixel.R, background.R)
	pixel.G = uncomposite(pixel.G, background.G)
	pixel.B = uncomposite(pixel.B, background.B)
	return pixel
}

func scrubTransparentRGB(source *image.NRGBA) {
	width := source.Bounds().Dx()
	height := source.Bounds().Dy()
	for y := range height {
		for x := range width {
			if source.NRGBAAt(x, y).A == 0 {
				source.SetNRGBA(x, y, color.NRGBA{})
			}
		}
	}
}

func colorDistance(left color.NRGBA, right color.NRGBA) int {
	return max(
		abs(int(left.R)-int(right.R)),
		abs(int(left.G)-int(right.G)),
		abs(int(left.B)-int(right.B)),
	)
}

func Clone(source *image.NRGBA) *image.NRGBA {
	result := image.NewNRGBA(image.Rect(0, 0, source.Bounds().Dx(), source.Bounds().Dy()))
	draw.Draw(result, result.Bounds(), source, source.Bounds().Min, draw.Src)
	return result
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
