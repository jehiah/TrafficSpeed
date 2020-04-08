package labelimg

import (
	"fmt"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prettyGrey(g *image.Gray) string {
	return prettyPix(g.Pix, g.Rect, g.PixOffset)
}
func prettyPaletted(g *image.Paletted) string {
	return prettyPix(g.Pix, g.Rect, g.PixOffset)
}

func prettyPix(pix []uint8, r image.Rectangle, pixOffset func(int, int) int) string {
	var s string = "\n"
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			s += fmt.Sprintf("%02x", pix[pixOffset(x, y)])
		}
		s += "\n"
	}
	return s
}

func TestLabelimg(t *testing.T) {
	type testCase struct {
		Rect             image.Rectangle
		Pix              []uint8
		Expect           []uint8
		ContiguousPixels int
		MinPixels        int
	}

	tests := []testCase{
		{
			Rect: image.Rect(1, 1, 5, 3), //  4 x 2
			Pix: []uint8{
				3, 3, 0, 4,
				3, 0, 4, 4,
			},
			Expect: []uint8{
				1, 1, 0, 2,
				1, 0, 2, 2,
			},
			ContiguousPixels: 1,
			MinPixels:        1,
		},
		{
			Rect: image.Rect(0, 0, 4, 4),
			Pix: []uint8{
				3, 3, 0, 4,
				3, 0, 0, 4,
				0, 0, 4, 4,
				0, 0, 4, 4,
			},
			Expect: []uint8{
				0, 0, 0, 1,
				0, 0, 0, 1,
				0, 0, 1, 1,
				0, 0, 1, 1,
			},
			ContiguousPixels: 1,
			MinPixels:        3,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			start := &image.Gray{
				Pix:    tc.Pix,
				Rect:   tc.Rect,
				Stride: tc.Rect.Dx(),
			}
			t.Logf("start %s", prettyGrey(start))

			labeled := New(start, tc.ContiguousPixels, tc.MinPixels)
			t.Logf("labeled %s", prettyPaletted(labeled))
			start.Pix = tc.Expect
			t.Logf("Expect %s", prettyGrey(start))
			assert.Equal(t, tc.Expect, labeled.Pix)
		})
	}
}
