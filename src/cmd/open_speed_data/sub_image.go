package main

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

// a delta image is a - b on a greyscale value of the total absolute combined r+g+b difference
func SubImage(a, b *image.RGBA) *image.Gray {
	if a.Bounds() != b.Bounds() {
		panic("not same size")
	}
	t := image.NewGray(a.Bounds())
	for i := 0; i*4 < len(a.Pix); i++ {
		r := abs16(a.Pix[(i*4)], b.Pix[(i*4)])
		g := abs16(a.Pix[(i*4)+1], b.Pix[(i*4)+1])
		b := abs16(a.Pix[(i*4)+2], b.Pix[(i*4)+2])
		sum := (r + g + b) / 3 // max delta = 0-255 * 3
		// clamp [0, 255]
		switch {
		case sum > 255:
			t.Pix[i] = 255
		default:
			t.Pix[i] = uint8(sum)
		}
	}
	return t
}
