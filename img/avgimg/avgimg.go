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
	"sort"
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

type AvgRGBA struct {
	Images []*image.RGBA
}

func (a AvgRGBA) ColorModel() color.Model {
	return color.RGBAModel
}
func (a AvgRGBA) Bounds() image.Rectangle {
	return a.Images[0].Bounds()
}
func (a AvgRGBA) At(x, y int) color.Color {
	var ar, ag, ab uint64
	for _, p := range a.Images {
		offset := p.PixOffset(x, y)
		ar += uint64(p.Pix[offset])
		ag += uint64(p.Pix[offset+1])
		ab += uint64(p.Pix[offset+2])
		// aa += uint64(p.Pix[offset+3])
	}
	l := uint64(len(a.Images))
	c := color.RGBA{
		uint8(ar / l),
		uint8(ag / l),
		uint8(ab / l),
		255,
	}

	if x <= 1 && y <= 1 {
		log.Printf("x:%d y:%d color:%v", x, y, c)
	}
	return c
}

func (a *AvgRGBA) Add(i *image.RGBA) {
	if len(a.Images) == 0 {
		a.Images = append(a.Images, i)
		return
	}
	if a.Images[0].Bounds() != i.Bounds() {
		panic("image bounds don't match")
	}
	a.Images = append(a.Images, i)
}

func (a *AvgRGBA) Size() int {
	return len(a.Images)
}

func (a AvgRGBA) Image() *image.RGBA {
	ab := a.Bounds()
	i := image.NewRGBA(image.Rect(0, 0, ab.Dx(), ab.Dy()))
	draw.Draw(i, i.Bounds(), a, ab.Min, draw.Over)
	return i
}

type MedianRGBA struct {
	AvgRGBA
}

func (a MedianRGBA) At(x, y int) color.Color {
	l := len(a.Images)
	median := l / 2
	var ar, ag, ab []uint8 = make([]uint8, l), make([]uint8, l), make([]uint8, l)
	for i, p := range a.Images {
		offset := p.PixOffset(x, y)
		ar[i] = p.Pix[offset]
		ag[i] = p.Pix[offset+1]
		ab[i] = p.Pix[offset+2]
	}
	sort.Slice(ar, func(i, j int) bool { return ar[i] < ar[j] })
	sort.Slice(ag, func(i, j int) bool { return ag[i] < ag[j] })
	sort.Slice(ab, func(i, j int) bool { return ab[i] < ab[j] })

	c := color.RGBA{
		ar[median],
		ag[median],
		ab[median],
		255,
	}
	return c
}

func (a MedianRGBA) Image() *image.RGBA {
	ab := a.Bounds()
	i := image.NewRGBA(image.Rect(0, 0, ab.Dx(), ab.Dy()))
	draw.Draw(i, i.Bounds(), a, ab.Min, draw.Over)
	return i
}
