// avgimg computes an "average" image given multiple Images
//
// This gives a simple way to average out noise from multiple frames and identify the
// static background separate from any foreground motion in a frame
package avgimg

import (
	"image"
	"image/color"
	"image/draw"
	"log"
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
	var ar, ag, ab uint64
	for _, p := range a {
		offset := p.PixOffset(x, y)
		ar += uint64(p.Pix[offset])
		ag += uint64(p.Pix[offset+1])
		ab += uint64(p.Pix[offset+2])
		// aa += uint64(p.Pix[offset+3])
	}
	c := color.RGBA{
		uint8(ar / uint64(len(a))),
		uint8(ag / uint64(len(a))),
		uint8(ab / uint64(len(a))),
		255,
	}

	if x <= 1 && y <= 1 {
		log.Printf("x:%d y:%d color:%v", x, y, c)
	}
	return c
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
	ab := a.Bounds()
	log.Printf("Image() %#v", ab)
	i := image.NewRGBA(image.Rect(0, 0, ab.Dx(), ab.Dy()))
	draw.Draw(i, i.Bounds(), a, ab.Min, draw.Over)
	return i
}
