package main

import (
	"image"
	"image/draw"
)

// CopyYCbCr returns a deep copy of i
//
// This is useful when the backing slice needs to be copied
func CopyYCbCr(i image.YCbCr) *image.YCbCr {
	f := image.NewYCbCr(i.Bounds(), i.SubsampleRatio)
	copy(f.Y, i.Y)
	copy(f.Cb, i.Cb)
	copy(f.Cr, i.Cr)
	return f
}

func RGBA(src *image.YCbCr) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, src.Rect.Dx(), src.Rect.Dy()))
	draw.Draw(m, m.Bounds(), src, image.ZP, draw.Src)
	return m
}

// grey scale
// for i := 0; i < len(vf.Image.Cb); i++ {
// 	vf.Image.Cb[i] = 128 // aka .5 the zero point
// }
// for i := 0; i < len(vf.Image.Cr); i++ {
// 	vf.Image.Cr[i] = 128
// }
// vf.Image.Cb = make([]uint8, len(vf.Image.Cb))
// vf.Image.Cr = make([]uint8, len(vf.Image.Cr))
