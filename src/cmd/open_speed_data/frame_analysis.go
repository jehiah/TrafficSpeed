package main

import (
	"html/template"
	"image"
	"log"
	"time"

	"blurimg"
	"diffimg"
	"github.com/nfnt/resize"
	"imgutils"
	"labelimg"
)

const analysFrameCount = 60

var analyizeInterval time.Duration = time.Duration(30) * time.Second

type FrameAnalysis struct {
	Timestamp time.Duration `json:"ts"`

	Base         template.URL `json:"base,omitempty"` // first frame
	BaseGif      template.URL `json:"base_gif,omitempty"`
	Highlight    template.URL `json:"highlight,omitempty"` // frame - bg
	HighlightGif template.URL `json:"highlight_gif,omitempty"`
	Colored      template.URL `json:"colored,omitempty"` // detected items in highlight
	ColoredGif   template.URL `json:"colored_gif,omitempty"`

	Positions []labelimg.Label `json:"positions,omitempty"`

	images []*image.RGBA
}

func (f FrameAnalysis) NeedsMore() bool {
	return len(f.images) < analysFrameCount
}

func (f *FrameAnalysis) Calculate(bg *image.RGBA, blurRadius, contiguousPixels, minMass int, tolerance uint8) {
	log.Printf("analysis covering %d frames starting at %s", len(f.images), f.Timestamp)
	if len(f.images) == 0 {
		return
	}
	src := f.images[0]
	f.Base = dataImg(src, "image/png")
	highlight := diffimg.DiffRGBA(src, bg, diffimg.SumDifferenceCap)
	f.Highlight = dataImg(highlight, "image/png")
	imgutils.BWClamp(highlight, tolerance)
	highlight = blurimg.Blur(highlight, blurRadius)
	detected := labelimg.New(highlight, contiguousPixels, minMass)
	f.Colored = dataImg(detected, "image/png")
	f.Positions = labelimg.Labels(detected)
	log.Printf("%#v", f.Positions)

	// animate 2s of video in 1s (drop 50% of frames)
	// ^^ == highlight_gif
	log.Printf("animating highlight gif")
	var baseGif []image.Image
	var highlightGif []image.Image
	var coloredGif []image.Image

	// bgresize := resize.Thumbnail(550, 200, bg, resize.NearestNeighbor).(*image.RGBA)
	for i := 0; i < 60 && i < len(f.images); i += 3 {
		im := f.images[i]
		baseGif = append(baseGif, resize.Thumbnail(550, 200, im, resize.NearestNeighbor))
		detected := diffimg.DiffRGBA(im, bg, diffimg.SumDifferenceCap)
		imgutils.BWClamp(detected, tolerance)
		detected = blurimg.Blur(detected, blurRadius)

		highlightGif = append(highlightGif, resize.Thumbnail(550, 200, detected, resize.NearestNeighbor))
		// since operating on a scaled image, we need to scale contiguousPixels and minMass here
		colored := labelimg.New(detected, contiguousPixels, minMass)
		coloredGif = append(coloredGif, resize.Thumbnail(550, 200, colored, resize.NearestNeighbor))
	}
	f.BaseGif = dataGif(imgutils.NewGIF(baseGif))
	f.HighlightGif = dataGif(imgutils.NewGIF(highlightGif))
	f.ColoredGif = dataGif(imgutils.NewGIF(coloredGif))
}
