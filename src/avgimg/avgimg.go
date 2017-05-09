// avgimg provides implements an Image which exposes the average values of multiple image.YCbCr
//
// This in essense gives a simple way to average out noise from multiple frames and identify the
// static background separate from any foreground motion in a frame
package avgimg

import (
	"image"
	"image/color"
	"image/draw"
)

type AvgImage []*image.YCbCr

func (a AvgImage) ColorModel() color.Model {
	return color.YCbCrModel
}
func (a AvgImage) Bounds() image.Rectangle {
	return a[0].Bounds()
}
func (a AvgImage) At(x, y int) color.Color {
	var ay, ab, ar uint64
	for _, p := range a {
		yi := p.YOffset(x, y)
		ci := p.COffset(x, y)
		ay += uint64(p.Y[yi])
		ab += uint64(p.Cb[ci])
		ar += uint64(p.Cr[ci])
	}
	return color.YCbCr{
		uint8(ay / uint64(len(a))),
		uint8(ab / uint64(len(a))),
		uint8(ar / uint64(len(a))),
	}
}

func (a AvgImage) Add(i *image.YCbCr) {
	if len(a) == 0 {
		a = append(a, i)
		return
	}
	if a[0].Bounds() != i.Bounds() {
		panic("image bounds don't match")
	}
	a = append(a, i)
}

func (a AvgImage) Image() *image.RGBA {
	i := image.NewRGBA(a.Bounds())
	draw.Draw(i, a.Bounds(), a, image.ZP, draw.Over)
	return i
}
