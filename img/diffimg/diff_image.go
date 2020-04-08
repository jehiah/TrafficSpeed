// diffimg computes the delta (in greyscale) of two color images
package diffimg

import (
	"image"
)

type Mode int

const (
	MaxDifference    Mode = iota // max of r, g or b diff
	SumDifference                // 1/3 of r+g+b diff
	SumDifferenceCap             // sum of r+g+b diff capped at 255
	MultDifference
	// YCbCrDifference

)

func absdiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	if a < b {
		return b - a
	}
	return 0
}

// DiffRGBA computes an image with the delta a - b in a greyscale image.
// the delta is the combined absolute r+g+b difference for each pixel
// The result is converted to black / white provided thresholdvalue
// a and b must have the same width and height
func DiffRGBA(a, b *image.RGBA, mode Mode) *image.Gray {
	aMin := a.Bounds().Min
	bMin := b.Bounds().Min
	dx, dy := a.Bounds().Dx(), a.Bounds().Dy()

	if dx != b.Bounds().Dx() || dy != b.Bounds().Dy() {
		panic("not same size")
	}
	gg := image.NewGray(image.Rect(0, 0, dx, dy))

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			aOffset := a.PixOffset(aMin.X+x, aMin.Y+y)
			bOffset := b.PixOffset(bMin.X+x, bMin.Y+y)

			r := absdiff(a.Pix[aOffset], b.Pix[bOffset])
			g := absdiff(a.Pix[aOffset+1], b.Pix[bOffset+1])
			b := absdiff(a.Pix[aOffset+2], b.Pix[bOffset+2])
			// max delta = 0-255 * 3
			// sum := (r + g + b) / 3
			switch mode {
			case MaxDifference:
				gg.Pix[gg.PixOffset(x, y)] = max(r, g, b)
			case SumDifference:
				gg.Pix[gg.PixOffset(x, y)] = uint8((uint16(r) + uint16(g) + uint16(b)) / 3)
			case SumDifferenceCap:
				delta := uint16(r) + uint16(g) + uint16(b)
				if delta > 255 {
					gg.Pix[gg.PixOffset(x, y)] = 255
				} else {
					gg.Pix[gg.PixOffset(x, y)] = uint8(delta)
				}
			case MultDifference:
				delta := (uint16(r) * uint16(g) * uint16(b)) / 1000
				if delta > 255 {
					gg.Pix[gg.PixOffset(x, y)] = 255
				} else {
					gg.Pix[gg.PixOffset(x, y)] = uint8(delta)
				}
			}
		}
	}
	return gg
}

func max(v ...uint8) (max uint8) {
	if len(v) <= 1 {
		panic("insufficient arguments")
	}
	for _, n := range v {
		if n > max {
			max = n
		}
	}
	return
}
