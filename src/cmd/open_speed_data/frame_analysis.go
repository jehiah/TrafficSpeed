package main

import (
	"html/template"
	"image"
	"image/draw"
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
	// YCbCr -> RGBA
	src := f.images[0]
	b := src.Bounds()
	m0 := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m0, m0.Bounds(), src, image.ZP, draw.Src)

	// // YCbCr -> RGBA
	// src = f.images[1]
	// b = src.Bounds()
	// m1 := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	// draw.Draw(m1, m1.Bounds(), src, image.ZP, draw.Src)

	highlight := SubImage(m0, bg)
	// base = first frame
	// highlight = base - background
	f.Base = dataImg(f.images[0], "image/png")
	f.Highlight = dataImg(highlight, "image/png")

	// animate 2s of video in 1s (drop 50% of frames)
	// ^^ == highlight_gif
}
