// labelimg creates a image.Palette that detects up to 255 non-overlapping areas
// in an image and assigns each one a unique color
package labelimg

import (
	"image"
	"image/color"
	"log"
)

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

// type LabelImage image.Paletted

// New creates a new paletted image by detecting contiguous blobs of non-zero color in `g`.
// overlap is defined as +/- contiguous pixes on x and y axis. Diagonal overlap is detected for
// contiguousPixels > 1. Groups of pixels smaller than minPixels are not returned
func New(g *image.Gray, contiguousPixels, minPixels int) *image.Paletted {
	pb := image.Rect(0, 0, g.Bounds().Dx(), g.Bounds().Dy())
	// log.Printf("new image %v", pb)
	p := image.NewPaletted(pb, color.Palette{color.Gray{0}})
	if contiguousPixels < 1 {
		panic("negative distance not allowed")
	}
	minOffset, maxOffset := -1*contiguousPixels, contiguousPixels

	for x := g.Rect.Min.X; x < g.Rect.Max.X; x++ {
		for y := g.Rect.Min.Y; y < g.Rect.Max.Y; y++ {
			o := g.PixOffset(x, y)
			// log.Printf("(%d,%d)[%d] = %d", x, y, o, g.Pix[0])
			if g.Pix[o] == 0 {
				continue
			}
			var i uint8
			// do overlap checks to see if this is a new point or if it overlaps
			for xo := minOffset; xo <= maxOffset; xo++ {
				for yo := minOffset; yo <= maxOffset; yo++ {
					if abs(xo)+abs(yo) > maxOffset {
						continue
					}
					if xo == 0 && yo == 0 {
						continue
					}

					xx, yy := x+xo, y+yo
					// log.Printf("checking offset %d,%d +/- %d/%d -> %d,%d", x, y, xo, yo, xx, yy)
					if !(image.Point{xx, yy}.In(g.Rect)) {
						// log.Printf("point (%d,%d) doesn't exist in %v", xx, yy, g.Rect)
						continue
					}
					if g.Pix[g.PixOffset(xx, yy)] != 0 {
						// contiguous area; check for existing detection
						oi := p.ColorIndexAt(xx-g.Rect.Min.X, yy-g.Rect.Min.Y)
						// log.Printf("contiguous area %d,%d = %d", xx, yy, oi)
						if oi == 0 {
							// that point will be detected later
							// log.Printf("%d,%d will be detected later. skipping", xx, yy)
							continue
						}
						if i == 0 {
							// log.Printf("%d,%d (i:0) matches existing detection at %d,%d of %d", x, y, xx, yy, oi)
							i = oi
						} else if i == oi {
							// log.Printf("%d,%d (i:%d) matches existing detection at %d,%d of %d", x, y, i, xx, yy, oi)
						} else if i != oi {
							// log.Printf("overlapping colors at (%d, %d) color index i:%v oi:%v", xx, yy, i, oi)
							// we need to treat these indexes as the same
							// use the smaller of the two
							if oi > i {
								replaceColor(p, oi, i)
							} else {
								replaceColor(p, i, oi)
								i = oi
							}
						}
					}
				}
			}
			if i == 0 {
				// add new color
				if len(p.Palette) == 254 {
					log.Printf("skipping detection of x,y (%d,%d); over 255 max", x, y)
					continue
				}
				i = uint8(len(p.Palette))
				// log.Printf("%d,%d **** NEW COLOR %d", x, y, i)
				p.Palette = append(p.Palette, color.Gray{i})
			}
			p.SetColorIndex(x-g.Rect.Min.X, y-g.Rect.Min.Y, i)
		}
	}

	// detect groups larger than minPixels
	if minPixels > 1 {
		var hits [255]int
		for _, pixel := range p.Pix {
			hits[pixel]++
		}
		// log.Printf("hits %#v threshold %d", hits, minPixels)
		// walk in reverse
		for n := 254; n > 0; n-- {
			if hits[n] > 0 && hits[n] <= minPixels {
				// log.Printf("zeroing pallete %d for %d (under threshold %d)", n, hits[n], minPixels)
				replaceColor(p, uint8(n), 0)
			}
		}
	}

	p.Palette = Glasbey
	return p
}

// replace pallete index a with b in p. Everything >=a is shifted down one index
// b must be < a
func replaceColor(p *image.Paletted, a, b uint8) {
	// log.Printf("replacing color %d with %d", a, b)
	if a <= b {
		panic("a palette index must be > b")
	}
	for i, c := range p.Pix {
		if c == a {
			p.Pix[i] = b
		} else if c > a {
			// shift down by one
			p.Pix[i] = c - 1
		}
	}
	// remove index a
	np := p.Palette[:a]
	if len(p.Palette) > int(a) {
		np = append(np, p.Palette[a+1:]...)
	}
	p.Palette = np
}
