package main

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
		pi := image.NewPaletted(im.Bounds(), p)
		draw.FloydSteinberg.Draw(pi, im.Bounds(), im, image.ZP)
		g.Image = append(g.Image, pi)
		g.Delay = append(g.Delay, 10)
	}
	return g
}
