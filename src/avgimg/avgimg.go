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

type AvgYCbCr []*image.YCbCr

func (a AvgYCbCr) ColorModel() color.Model {
	return color.YCbCrModel
}
func (a AvgYCbCr) Bounds() image.Rectangle {
	return a[0].Bounds()
}
func (a AvgYCbCr) At(x, y int) color.Color {
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

func (a AvgYCbCr) Add(i *image.YCbCr) {
	if len(a) == 0 {
		a = append(a, i)
		return
	}
	if a[0].Bounds() != i.Bounds() {
		panic("image bounds don't match")
	}
	a = append(a, i)
}

func (a AvgYCbCr) Image() *image.RGBA {
	i := image.NewRGBA(a.Bounds())
	draw.Draw(i, a.Bounds(), a, image.ZP, draw.Over)
	return i
}

type AvgRGBA []*image.RGBA

func (a AvgRGBA) ColorModel() color.Model {
	return color.RGBAModel
}
func (a AvgRGBA) Bounds() image.Rectangle {
	return a[0].Bounds()
}
func (a AvgRGBA) At(x, y int) color.Color {
	var ar, ag, ab, aa uint64
	for _, p := range a {
		offset := (y * p.Stride) + (x * 4)
		ar += uint64(p.Pix[offset])
		ag += uint64(p.Pix[offset+1])
		ab += uint64(p.Pix[offset+2])
		aa += uint64(p.Pix[offset+3])
	}
	return color.RGBA{
		uint8(ar / uint64(len(a))),
		uint8(ag / uint64(len(a))),
		uint8(ab / uint64(len(a))),
		uint8(aa / uint64(len(a))),
	}
}

func (a AvgRGBA) Add(i *image.RGBA) {
	if len(a) == 0 {
		a = append(a, i)
		return
	}
	if a[0].Bounds() != i.Bounds() {
		panic("image bounds don't match")
	}
	a = append(a, i)
}

func (a AvgRGBA) Image() *image.RGBA {
	i := image.NewRGBA(a.Bounds())
	draw.Draw(i, a.Bounds(), a, image.ZP, draw.Over)
	return i
}
