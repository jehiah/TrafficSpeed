// labelimg creates a image.Pallete that detects up to 255 non-overlapping objects
// in an image and assigns each one a unique color
package labelimg

import (
	"image"
	"image/color"
	"log"
)

type LabelImage image.Paletted


func New(g *image.Gray) *image.Paletted {
	p := image.NewPaletted(g.Bounds(), nil)
	for x := g.Rect.Min.X; x < g.Rect.Max.X; x++ {
		for y := g.Rect.Min.Y; y < g.Rect.Max.Y; y++ {
			o := g.PixOffset(x, y)
			if g.Pix[o] == 0 {
				continue
			}
			var i uint8
			// do overlap checks to see if this is a new point or if it overlaps
			for _, offset := range [][2]int{{0,1},{0,-1}, {1, 0}, {-1, 0}} {
				xo, yo := offset[0], offset[1]
				xx, yy := x + xo, y + yo
				if !(image.Point{xx, yy}.In(p.Rect)) { 
					continue
				}
				if g.Pix[g.PixOffset(xx, yy)] != 0 {
					// lower index
					oi := p.ColorIndexAt(xx-g.Rect.Min.X, yy-g.Rect.Min.Y)
					if oi == 0 {
						// that point will be detected later
						continue
					}
					if i == 0 {
						i = oi
					} else if i != oi {
						log.Printf("overlapping colors at (%d, %d) color index i:%v oi:%v", xx, yy, i, oi)
						// we need to treat these indexes as the same
						// use the smaller of the two
					}
				}
			}
			if i == 0 {
				// add new color
				i = uint8(len(p.Palette))
				p.Palette = append(p.Palette, color.Gray{i})
			}
			p.SetColorIndex(x-g.Rect.Min.X, y-g.Rect.Min.Y, i)
		}
	}
	return p
}