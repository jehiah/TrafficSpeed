package main

// #cgo LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib"
// #cgo CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include"

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"time"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
	"github.com/nfnt/resize"
	// "github.com/nareix/joy4/format/mp4"
	"avgimg"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func init() {
	format.RegisterAll()
}

const bgFrameCount = 30

type Project struct {
	Err      error  `json:"error,omitempty"`
	Filename string `json:"filename"`

	// User Inputs
	Rotate       float64        `json:"rotate,omitempty"` // radians
	BBox         *BBox          `json:"bbox,omitempty"`
	Masks        []Mask         `json:"masks,omitempty"`
	Tolerance    float64        `json:"tolerance"`
	Blur         int64          `json:"blur"`
	MinMass      int64          `json:"min_mass"`
	Seek         float64        `json:"seek"`
	Calibrations []*Calibration `json:"calibrations"`

	Duration        time.Duration `json:"duration_seconds,omitempty"`
	VideoResolution string        `json:"video_resolution,omitempty"`
	Frames          int64         `json:"frames,omitempty"`

	Step     int      `json:"step"`
	Response Response `json:"response,omitempty"`

	demuxer av.DemuxCloser
}

type Response struct {
	Err               string          `json:"err,omitempty"`
	CroppedResolution string          `json:"cropped_resolution,omitempty"`
	OverviewGif       template.URL    `json:"overview_gif,omitempty"`
	OverviewImg       template.URL    `json:"overview_img,omitempty"`
	Step2Img          template.URL    `json:"step_2_img,omitempty"` // rotation
	Step3Img          template.URL    `json:"step_3_img,omitempty"` // crop
	Step4Img          template.URL    `json:"step_4_img,omitempty"` // mask
	Step4MaskImg      template.URL    `json:"step_4_mask_img,omitempty"`
	BackgroundImg     template.URL    `json:"background_img,omitempty"`
	FrameAnalysis     []FrameAnalysis `json:"frame_analysis,omitempty"`
	Step6Img          template.URL    `json:"step_6_img,omitempty"`
}

func NewProject(f string) *Project {
	// overview_gif
	// overview_img
	// duration_seconds
	// frames

	// Frames            int64           `json:"frames,omitempty"`
	// Duration          float64         `json:"duration_seconds,omitempty"`

	file, err := avutil.Open(f)
	if err != nil {
		log.Panicf("%s", err)
	}
	streams, err := file.Streams()
	if err != nil {
		log.Panicf("%s", err)
	}
	vstream := streams[0].(av.VideoCodecData)

	return &Project{
		Filename:        f,
		VideoResolution: fmt.Sprintf("%dx%d", vstream.Width(), vstream.Height()),

		demuxer: file,
	}
}

func (p *Project) Close() {
	if p.demuxer != nil {
		p.demuxer.Close()
		p.demuxer = nil
	}
}

func (p *Project) Run() error {
	p.SetStep()

	streams, _ := p.demuxer.Streams()
	decoders := make([]*ffmpeg.VideoDecoder, len(streams))
	for i, stream := range streams {
		vstream := stream.(av.VideoCodecData)
		var err error
		decoders[i], err = ffmpeg.NewVideoDecoder(vstream)
		if err != nil {
			return err
		}
	}

	// set overview img
	frame := 0
	var bg avgimg.AvgImage
	var err error
	for ; ; frame++ {
		var pkt av.Packet
		if pkt, err = p.demuxer.ReadPacket(); err != nil {
			if err != io.EOF {
				log.Printf("readPacket err: %s", err)
			}
			break
		}
		if !streams[pkt.Idx].Type().IsVideo() {
			continue
		}
		interested := true
		switch {
		case frame == 0:
		case p.Step == 5 && len(bg) < bgFrameCount && frame%15 == 0:
		default:
			interested = false
		}
		var vf *ffmpeg.VideoFrame
		if interested {
			log.Printf("interested in frame %d %s", frame, pkt.Time)
			// set overview image
			decoder := decoders[pkt.Idx]
			vf, err = decoder.Decode(pkt.Data)
			if err != nil {
				return err
			}
		}
		if frame == 0 {
			overview := resize.Thumbnail(400, 300, &vf.Image, resize.NearestNeighbor)
			p.Response.OverviewImg, err = dataImg(overview)
			if err != nil {
				return err
			}
			if p.Step == 2 {
				p.Response.Step2Img, err = dataImg(&vf.Image)
			}
			mw := imagick.NewMagickWand()
			background := imagick.NewPixelWand()
			background.SetColor("#000000")

			// load image
			out := new(bytes.Buffer)
			// grey scale
			for i := 0; i < len(vf.Image.Cb); i++ {
				vf.Image.Cb[i] = 128 // aka .5 the zero point
			}
			for i := 0; i < len(vf.Image.Cr); i++ {
				vf.Image.Cr[i] = 128
			}
			// vf.Image.Cb = make([]uint8, len(vf.Image.Cb))
			// vf.Image.Cr = make([]uint8, len(vf.Image.Cr))
			png.Encode(out, &vf.Image)
			mw.ReadImageBlob(out.Bytes())

			if p.Step >= 3 {
				log.Printf("rotating %v", p.Rotate)
				// apply rotation
				err = mw.RotateImage(background, RadiansToDegrees(p.Rotate))
				if err != nil {
					return err
				}
				mw.SetImageFormat("PNG")
				imgBytes := mw.GetImageBlob()
				p.Response.Step3Img = dataImgFromBytes(imgBytes)
			}
			if p.Step >= 4 {
				// rotate & crop
				log.Printf("crop %v", p.BBox)
				err = mw.CropImage(uint(p.BBox.Width()), uint(p.BBox.Height()), int(p.BBox.A.X), int(p.BBox.A.Y))
				if err != nil {
					return err
				}
				imgBytes := mw.GetImageBlob()
				p.Response.Step4Img = dataImgFromBytes(imgBytes)
			}
		}
		if p.Step == 5 && len(bg) < bgFrameCount && frame%15 == 0 {
			// calculate the background
			// background_img
			var bgframe image.YCbCr = vf.Image
			bg = append(bg, &bgframe)
		}

		// set every frame, so this ends w/ the last value
		p.Duration = pkt.Time
		p.Frames = int64(frame)
	}
	var bgavg *image.RGBA
	if p.Step == 5 && len(bg) > 0 {
		log.Printf("calculate background from %d frames", len(bg))
		bgavg = image.NewRGBA(bg.Bounds())
		draw.Draw(bgavg, bgavg.Bounds(), bg, image.ZP, draw.Over)
		p.Response.BackgroundImg, err = dataImg(bgavg)
		if err != nil {
			return err
		}
	}
	if p.Step == 5 {
		// pick 4 slices from the video
		// ts
		// base = first frame
		// highlight = base - background
		// animate 2s of video in 1s (drop 50% of frames)
		// ^^ == highlight_gif
	}

	return nil
}
