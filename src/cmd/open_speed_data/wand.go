package main

import (
	"fmt"
	"image"
	"log"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func WandSetImage(mw *imagick.MagickWand, src *image.RGBA) error {
	// LogWandInfo(mw, "[WandSetImage (before)]")
	cols, rows := uint(src.Rect.Dx()), uint(src.Rect.Dy())
	log.Printf("loading image %dx%d", cols, rows)
	var err error
	// err = mw.SetPage(cols, rows, 0, 0)
	// if err != nil {
	// 	log.Printf("err SetPage %s", err)
	// }
	// err = mw.SetSize(cols, rows)
	// if err != nil {
	// 	log.Printf("err SetSize %s", err)
	// }
	// err := mw.SetImageExtent(cols, rows)
	// if err != nil {
	// 	return err
	// }
	// SetSizeOffset
	// SetPage
	// SetImagePage
	// LogWandInfo(mw, "[WandSetImage (before ConstituteImage)]")
	err = mw.ConstituteImage(cols, rows, "RGBA", imagick.PIXEL_CHAR, src.Pix)
	// LogWandInfo(mw, "[WandSetImage (after ConstituteImage)]")
	err = mw.SetImagePage(cols, rows, 0, 0)
	if err != nil {
		log.Printf("err SetSize %s", err)
	}
	return err
	// return mw.ImportImagePixels(0, 0, cols, rows, "RGBA", imagick.PIXEL_CHAR, src.Pix)
}

func LogWandInfo(mw *imagick.MagickWand, prefix string) {
	pw, ph, px, py, err := mw.GetImagePage()
	log.Printf("%sGetImagePage: %v %v %v %v %s", prefix, pw, ph, px, py, err)
	// pw, ph, px, py, err := mw.GetPage()
	// log.Printf("%sGetPage: %v %v %v %v %s", prefix, pw, ph, px, py, err)
	// ww, hh, err := mw.GetSize()
	// log.Printf("%sGetSize: %v %v %s", prefix, ww, hh, err)
	// o, err := mw.GetSizeOffset()
	// log.Printf("%sGetSizeOffset: %v %s", prefix, o, err)
	// rx, ry, err := mw.GetResolution()
	// log.Printf("%sGetResolution: %v %v %s ", prefix, rx, ry, err)
}

func WandImage(mw *imagick.MagickWand) (*image.RGBA, error) {
	LogWandInfo(mw, "[WandImage]")
	width, height, _, _, _ := mw.GetImagePage()
	ww, hh, _ := mw.GetSize()
	if ww > width {
		width = ww
	}
	if hh > height {
		height = hh
	}
	if width == 0 || height == 0 {
		return nil, fmt.Errorf("width or height == 0; %v %v", width, height)
	}
	return WandImageSize(mw, image.Rect(0, 0, int(width), int(height)))
}

func WandImageSize(mw *imagick.MagickWand, r image.Rectangle) (*image.RGBA, error) {
	LogWandInfo(mw, fmt.Sprintf("[WandImageSize %#v]", r))
	data, err := mw.ExportImagePixels(r.Min.X, r.Min.Y, uint(r.Dx()), uint(r.Dy()), "RGBA", imagick.PIXEL_CHAR)
	if err != nil {
		return nil, err
	}
	return &image.RGBA{
		data.([]uint8),
		r.Dx() * 4,
		r,
	}, nil
}

func Crop(mw *imagick.MagickWand, r image.Rectangle) *imagick.MagickWand {
	_, _, wx, wy, _ := mw.GetImagePage()
	newSize := fmt.Sprintf("%dx%d", r.Dx(), r.Dy())
	log.Printf("crop to %d,%d %s", r.Min.X+wx, r.Min.Y+wy, newSize)
	mw = mw.GetImageRegion(uint(r.Dx()), uint(r.Dy()), r.Min.X+wx, r.Min.Y+wy)
	mw.ResetImagePage(newSize + "+0+0")
	return mw

	// err = mw.CropImage(uint(p.BBox.Width()), uint(p.BBox.Height()), int(p.BBox.A.X)+wx, int(p.BBox.A.Y)+wy)
	// if err != nil {
	// 	return err
	// }
}
