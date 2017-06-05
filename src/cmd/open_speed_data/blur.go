package main

import (
	"image"
)

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

// Blur provides a simple blur expecting a black & white image
func Blur(g *image.Gray, radius int) *image.Gray {
	if radius <= 0 {
		return g
	}
	gg := image.NewGray(g.Rect)
	var white uint8 = 255
	min, max := radius*-1, radius
	for x := 0; x < g.Rect.Dx(); x++ {
		for y := 0; y < g.Rect.Dy(); y++ {
			offset := g.PixOffset(x, y)
			if g.Pix[offset] != white {
				continue
			}
			for xo := min; xo <= max; xo++ {
				for yo := min; yo <= max; yo++ {
					if abs(xo)+abs(yo) > radius {
						continue
					}
					if (image.Point{x + xo, y + yo}).In(gg.Rect) {
						gg.Pix[gg.PixOffset(x+xo, y+yo)] = white
					}
				}
			}
		}
	}
	return gg
}
