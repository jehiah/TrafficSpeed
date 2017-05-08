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
	// "github.com/nareix/joy4/format/mp4"
	"avgimg"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func init() {
	format.RegisterAll()
}

const bgFrameCount = 30
const bgFrameSkip = 7

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

	DebugImages []template.URL
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

	var analysis = &FrameAnalysis{}

	// set overview img
	frame := 0
	var bg avgimg.AvgImage
	var err error
	mw := imagick.NewMagickWand()
	background := imagick.NewPixelWand()
	background.SetColor("#000000")

	for ; ; frame++ {
		var pkt av.Packet
		if pkt, err = p.demuxer.ReadPacket(); err != nil {
			if err != io.EOF {
				log.Printf("readPacket err: %s", err)
				return err
			}
			err = nil
			break
		}
		if !streams[pkt.Idx].Type().IsVideo() {
			continue
		}
		p.Duration = pkt.Time
		p.Frames = int64(frame)


		interested := true
		switch {
		case frame == 0:
		case p.Step == 5 && len(bg) < bgFrameCount:
			// get all frames until we have a background because frames are dependent on the previous frame
		case p.Step == 5 && analysis.NeedsMore():
		default:
			interested = false
		}

		var vf *ffmpeg.VideoFrame
		if interested {
			log.Printf("interested in frame %d %s", frame, pkt.Time)
			vf, err = decoders[pkt.Idx].Decode(pkt.Data)
			if err != nil {
				return err
			}
		}

		if frame == 0 {
			// set overview image
			p.Response.OverviewImg = dataImgWithSize(&vf.Image, 400, 300, "")
			if p.Step == 2 {
				p.Response.Step2Img = dataImg(&vf.Image, "")
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
			// load image
			
			out := new(bytes.Buffer)
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
				p.Response.Step3Img = dataImgFromBytes("image/png", imgBytes)
			}
			if p.Step >= 4 {
				// rotate & crop
				log.Printf("crop %v", p.BBox)
				err = mw.CropImage(uint(p.BBox.Width()), uint(p.BBox.Height()), int(p.BBox.A.X), int(p.BBox.A.Y))
				if err != nil {
					return err
				}
				imgBytes := mw.GetImageBlob()
				p.Response.Step4Img = dataImgFromBytes("image/png", imgBytes)
			}
		}

		var frameImg *image.YCbCr
		if p.Step == 5 && len(bg) < bgFrameCount && frame%bgFrameSkip == 0 {
			// queue frames to be used in background calculation
			if frameImg == nil {
				frameImg = CopyYCbCr(vf.Image)
			}
			bg = append(bg, frameImg)
			// debugImg := dataImgWithSize(bgframe, 400, 300, "")
			// p.Response.DebugImages = append(p.Response.DebugImages, frameImg)
		}
		if p.Step == 5 && analysis.NeedsMore() {
			if frameImg == nil {
				frameImg = CopyYCbCr(vf.Image)
			}
			log.Printf("saving frame %d for analysis later", frame)
			analysis.images = append(analysis.images, frameImg)
		}
		// set every frame, so this ends w/ the last value
	}
	
	var bgavg *image.RGBA
	if p.Step == 5 && len(bg) > 0 {
		log.Printf("calculate background from %d frames", len(bg))
		bgavg = image.NewRGBA(bg.Bounds())
		draw.Draw(bgavg, bgavg.Bounds(), bg, image.ZP, draw.Over)
		p.Response.BackgroundImg = dataImg(bgavg, "")
		
	}
	if p.Step == 5 && bgavg != nil{
		analysis.Calculate(bgavg)
		p.Response.FrameAnalysis = append(p.Response.FrameAnalysis, *analysis)
	}

	return nil
}
