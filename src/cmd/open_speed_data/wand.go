package main

import (
	"image"
	"log"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func WandSetImage(mw *imagick.MagickWand, src *image.RGBA) error {
	cols, rows := uint(src.Rect.Dx()), uint(src.Rect.Dy())
	log.Printf("loading image %dx%d", cols, rows)
	// err := mw.SetImagePage(cols, rows, 0, 0)
	err := mw.SetSize(cols, rows)
	if err != nil {
		return err
	}
	// err := mw.SetImageExtent(cols, rows)
	// if err != nil {
	// 	return err
	// }
	return mw.ConstituteImage(cols, rows, "RGBA", imagick.PIXEL_CHAR, src.Pix)
	// return mw.ImportImagePixels(0, 0, cols, rows, "RGBA", imagick.PIXEL_CHAR, src.Pix)
}

func WandImage(mw *imagick.MagickWand) (*image.RGBA, error) {
	width, height, x, y, err := mw.GetImagePage()
	log.Printf("GetImagePage: %v %v %v %v", width, height, x, y)
	if err != nil {
		return nil, err
	}
	ww, hh, err := mw.GetSize()
	log.Printf("GetSize: %v %v ", ww, hh)
	
	pw, ph, px, py, err := mw.GetPage()
	log.Printf("GetPage: %v %v %v %v", pw, ph, px, py, err)
	
	
	if ww > width {
		width = ww
	}
	if hh > height {
		height = hh
	}
	if err != nil {
		return nil, err
	}
	return WandImageSize(mw, image.Rect(0, 0, int(width), int(height)))
}

func WandImageSize(mw *imagick.MagickWand, r image.Rectangle) (*image.RGBA, error) {
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
