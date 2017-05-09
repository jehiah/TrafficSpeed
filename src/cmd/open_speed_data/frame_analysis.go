package main

import (
	"html/template"
	"image"
	"log"
	"time"
)

const analysFrameCount = 60

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

	images []*image.YCbCr
}

func (f FrameAnalysis) NeedsMore() bool {
	return len(f.images) < analysFrameCount
}

func (f *FrameAnalysis) Calculate(bg *image.RGBA) {
	log.Printf("analysis covering %d frames starting at %s", len(f.images), f.Timestamp)
	if len(f.images) == 0 {
		return
	}
	src := RGBA(f.images[0])
	highlight := SubImage(src, bg)
	// base = first frame
	// highlight = base - background
	f.Base = dataImg(f.images[0], "image/png")
	f.Highlight = dataImg(highlight, "image/png")

	// animate 2s of video in 1s (drop 50% of frames)
	// ^^ == highlight_gif
}
