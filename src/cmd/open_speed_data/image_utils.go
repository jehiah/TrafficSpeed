package main

import (
	"image"
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
