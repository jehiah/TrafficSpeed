package main

// #cgo LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib"
// #cgo CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include"

import (
	"bytes"
	"fmt"
	"html/template"
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
	"gopkg.in/gographics/imagick.v3/imagick"
)

func init() {
	format.RegisterAll()
}

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
	Step2Img          template.URL    `json:"step_2_img,omitempty"`
	Step3Img          template.URL    `json:"step_3_img,omitempty"`
	Step4Img          template.URL    `json:"step_4_img,omitempty"`
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
	for ; ; frame++ {
		var pkt av.Packet
		var err error
		if pkt, err = p.demuxer.ReadPacket(); err != nil {
			if err != io.EOF {
				log.Printf("readPacket err: %s", err)
			}
			break
		}
		if !streams[pkt.Idx].Type().IsVideo() {
			continue
		}
		if frame == 0 {
			// set overview image
			decoder := decoders[pkt.Idx]
			vf, err := decoder.Decode(pkt.Data)
			if err != nil {
				return err
			}
			overview := resize.Thumbnail(400, 300, &vf.Image, resize.NearestNeighbor)
			p.Response.OverviewImg, err = dataImg(overview)
			if err != nil {
				return err
			}
			if p.Step == 2 {
				p.Response.Step2Img, err = dataImg(&vf.Image)
			}
			if p.Step == 3 {
				// apply rotation
				mw := imagick.NewMagickWand()
				background := imagick.NewPixelWand()
				background.SetColor("#000000")
				out := new(bytes.Buffer)
				png.Encode(out, &vf.Image)
				mw.ReadImageBlob(out.Bytes())
				mw.RotateImage(background, RadiansToDegrees(p.Rotate))
				mw.SetImageFormat("PNG")
				imgBytes := mw.GetImageBlob()
				p.Response.Step3Img = dataImgFromBytes(imgBytes)
			}

		}
		p.Duration = pkt.Time
		p.Frames = int64(frame)

	}
	return nil
}
