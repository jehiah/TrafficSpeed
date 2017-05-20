package main

import (
	"github.com/nfnt/resize"
	"html/template"
	"image"
	"log"
	"time"
)

const analysFrameCount = 90

var analyizeInterval time.Duration = time.Duration(30) * time.Second

type FrameAnalysis struct {
	Timestamp    time.Duration `json:"ts"`
	Base         template.URL  `json:"base,omitempty"`
	BaseGif      template.URL  `json:"base_gif,omitempty"`
	Highlight    template.URL  `json:"highlight,omitempty"`
	HighlightGif template.URL  `json:"highlight_gif,omitempty"`
	Colored      template.URL  `json:"colored,omitempty"`
	ColoredGif   template.URL  `json:"colored_gif,omitempty"`
	Positions    []Position    `json:"positions,omitempty"`

	images []*image.RGBA
}

func (f FrameAnalysis) NeedsMore() bool {
	return len(f.images) < analysFrameCount
}

func (f *FrameAnalysis) Calculate(bg *image.RGBA, tolerance uint8) {
	log.Printf("analysis covering %d frames starting at %s", len(f.images), f.Timestamp)
	if len(f.images) == 0 {
		return
	}
	src := f.images[0]
	highlight := SubImage(src, bg, tolerance)
	// base = first frame
	// highlight = base - background
	f.Base = dataImg(f.images[0], "image/png")
	f.Highlight = dataImg(highlight, "image/png")

	// animate 2s of video in 1s (drop 50% of frames)
	// ^^ == highlight_gif
	log.Printf("animating highlight gif")
	var highlightImages []image.Image
	var colored []image.Image
	bgresize := resize.Thumbnail(550, 200, bg, resize.NearestNeighbor).(*image.RGBA)
	for i := 0; i < 60 && i < len(f.images); i += 3 {
		im := resize.Thumbnail(550, 200, f.images[i], resize.NearestNeighbor)
		highlightImages = append(highlightImages, im)

		colored = append(colored, SubImage(im.(*image.RGBA), bgresize, tolerance))
	}
	g := NewGIF(highlightImages)
	f.HighlightGif = dataGif(g)

	g = NewGIF(colored)
	f.ColoredGif = dataGif(g)

}
