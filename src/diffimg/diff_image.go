// diffimg computes the delta (in greyscale) of two color images
package diffimg

import (
	"image"
)

func abs16(a, b uint8) uint16 {
	if a > b {
		return uint16(a) - uint16(b)
	}
	if a < b {
		return uint16(b) - uint16(a)
	}
	return 0
}

// DiffRGBA computes an image with the delta a - b in a greyscale image.
// the delta is the combined absolute r+g+b difference for each pixel
// The result is converted to black / white provided thresholdvalue
// a and b must have the same width and height
func DiffRGBA(a, b *image.RGBA, tolerance uint8) *image.Gray {
	aMin := a.Bounds().Min
	bMin := b.Bounds().Min
	dx, dy := a.Bounds().Dx(), a.Bounds().Dy()

	if dx != b.Bounds().Dx() || dy != b.Bounds().Dy() {
		panic("not same size")
	}
	c := uint16(tolerance)
	gg := image.NewGray(image.Rect(0, 0, dx, dy))

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			aOffset := a.PixOffset(aMin.X+x, aMin.Y+y)
			bOffset := b.PixOffset(bMin.X+x, bMin.Y+y)
			r := abs16(a.Pix[aOffset], b.Pix[bOffset])
			g := abs16(a.Pix[aOffset+1], b.Pix[bOffset+1])
			b := abs16(a.Pix[aOffset+2], b.Pix[bOffset+2])
			// max delta = 0-255 * 3
			sum := (r + g + b) / 3
			// clamp 0/255
			if sum < c {
				gg.Pix[gg.PixOffset(x, y)] = 0
			} else {
				gg.Pix[gg.PixOffset(x, y)] = 255
			}
		}
	}
	return gg
}
