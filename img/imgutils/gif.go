package imgutils

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
)

func NewGIF(images []image.Image) *gif.GIF {
	p := palette.Plan9
	g := &gif.GIF{}
	for _, im := range images {
		if pi, ok := im.(*image.Paletted); ok {
			g.Image = append(g.Image, pi)
		} else {
			pi := image.NewPaletted(im.Bounds(), p)
			draw.FloydSteinberg.Draw(pi, im.Bounds(), im, image.ZP)
			g.Image = append(g.Image, pi)
		}
		g.Delay = append(g.Delay, 10)
	}
	return g
}
