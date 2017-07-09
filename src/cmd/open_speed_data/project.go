package main

// #cgo LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib"
// #cgo CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include"

import (
	"fmt"
	"html/template"
	"image"
	"io"
	"log"
	"time"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
	// "github.com/nareix/joy4/format/mp4"
	"avgimg"
	"github.com/anthonynsimon/bild/transform"
	"imgutils"
)

func init() {
	format.RegisterAll()
}

// roughly 7s of frames
const bgFrameCount = 15
const bgFrameSkip = 15

type Project struct {
	Err      error  `json:"error,omitempty"`
	Filename string `json:"filename"`

	// User Inputs
	PreCrop          *BBox          `json:"pre_crop,omitempty"`
	Rotate           float64        `json:"rotate,omitempty"` // radians
	PostCrop         *BBox          `json:"post_crop,omitempty"`
	Masks            Masks          `json:"masks,omitempty"`
	Tolerance        uint8          `json:"tolerance"`
	Blur             int            `json:"blur"`
	ContiguousPixels int            `json:"contiguous_pixels"`
	MinMass          int            `json:"min_mass"`
	Seek             float64        `json:"seek"`
	Calibrations     []*Calibration `json:"calibrations"`

	Duration        time.Duration `json:"duration_seconds,omitempty"`
	VideoResolution string        `json:"video_resolution,omitempty"`
	Frames          int64         `json:"frames,omitempty"`

	Step     int      `json:"step"`
	Response Response `json:"response,omitempty"`

	demuxer av.DemuxCloser
}

type Response struct {
	Err                  string          `json:"err,omitempty"`
	PreCroppedResolution string          `json:"pre_cropped_resolution,omitempty"`
	CroppedResolution    string          `json:"cropped_resolution,omitempty"`
	OverviewGif          template.URL    `json:"overview_gif,omitempty"`
	OverviewImg          template.URL    `json:"overview_img,omitempty"`
	Step2Img             template.URL    `json:"step_2_img,omitempty"` // rotation
	Step3Img             template.URL    `json:"step_3_img,omitempty"` // crop
	Step4Img             template.URL    `json:"step_4_img,omitempty"` // mask
	Step4MaskImg         template.URL    `json:"step_4_mask_img,omitempty"`
	BackgroundImg        template.URL    `json:"background_img,omitempty"`
	FrameAnalysis        []FrameAnalysis `json:"frame_analysis,omitempty"`
	FramePositions       []FramePosition
	Step6Img             template.URL `json:"step_6_img,omitempty"`

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
	if len(streams) == 0 {
		log.Fatalf("no streams in %s", f)
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
	start := time.Now()
	defer func() {
		log.Printf("Run took %s", time.Since(start))
	}()
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
	bg := &avgimg.MedianRGBA{}
	var bgavg *image.RGBA
	var err error
	analyzer := &Analyzer{
		BWCutoff:          p.Tolerance,
		BlurRadius:        p.Blur,
		ContinguousPixels: p.ContiguousPixels,
		MinMass:           p.MinMass,
	}

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
		case p.Step == 5 && len(bg.Images) < bgFrameCount:
			// get all frames until we have a background because frames are dependent on the previous frame
		case p.Step == 5 && analysis.NeedsMore():
		case p.Step == 6:
		default:
			interested = false
		}

		var vf *ffmpeg.VideoFrame
		var rgbImg *image.RGBA
		var img *image.RGBA
		if interested {
			log.Printf("interested in frame %d %s", frame, pkt.Time)
			vf, err = decoders[pkt.Idx].Decode(pkt.Data)
			if err != nil {
				return err
			}
			if vf == nil {
				log.Printf("no video frame?")
				frame--
				continue
			}
			rgbImg = imgutils.RGBA(&vf.Image)
			img = rgbImg
		}

		if frame == 0 {
			// set overview image
			if p.Step > 1 {
				p.Response.OverviewImg = dataImgWithSize(&vf.Image, 400, 300, "")
			} else {
				p.Response.OverviewImg = dataImg(&vf.Image, "image/png")
			}

			if p.PreCrop != nil {
				log.Printf("PreCrop %v", p.PreCrop)
				img = img.SubImage(p.PreCrop.Rect()).(*image.RGBA)
				p.Response.PreCroppedResolution = fmt.Sprintf("%dx%d", p.PreCrop.Dx(), p.PreCrop.Dy())
			}

			if p.Step == 2 {
				p.Response.Step2Img = dataImg(img, "")
			}

			if p.Step >= 3 {
				log.Printf("rotating %v", p.Rotate)
				// apply rotation

				img = transform.Rotate(img, RadiansToDegrees(p.Rotate), &transform.RotationOptions{ResizeBounds: true})
				p.Response.Step3Img = dataImg(img, "image/webp")
			}
			if p.Step >= 4 {
				// rotate & crop
				log.Printf("PostCrop %v", p.PostCrop)
				if p.PostCrop != nil {
					img = img.SubImage(p.PostCrop.Rect()).(*image.RGBA)
					// img = transform.Crop(img, p.BBox.Rect())
					p.Response.CroppedResolution = fmt.Sprintf("%dx%d", p.PostCrop.Dx(), p.PostCrop.Dy())
				} else {
					p.Response.CroppedResolution = fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy())
				}
				p.Response.Step4Img = dataImg(img, "image/webp")
			}
			if p.Step >= 5 {
				// mask
				p.Masks.Apply(img)
				p.Response.Step4MaskImg = dataImg(img, "image/webp")
			}
		}

		switch {
		case p.Step == 5 && len(bg.Images) < bgFrameCount && frame%bgFrameSkip == 0:
			fallthrough
		case p.Step == 5 && analysis.NeedsMore() || p.Step == 6:
			if p.PreCrop != nil {
				rgbImg = rgbImg.SubImage(p.PreCrop.Rect()).(*image.RGBA)
			}

			if p.Rotate != 0 {
				rgbImg = transform.Rotate(rgbImg, RadiansToDegrees(p.Rotate), &transform.RotationOptions{ResizeBounds: true})
			}
			if p.PostCrop != nil {
				rgbImg = rgbImg.SubImage(p.PostCrop.Rect()).(*image.RGBA)
			}
			p.Masks.Apply(rgbImg)
		}

		if p.Step >= 5 && len(bg.Images) < bgFrameCount && frame%bgFrameSkip == 0 {
			bg.Images = append(bg.Images, rgbImg)
			if len(bg.Images) == bgFrameCount {
				log.Printf("calculating background from %d frames", len(bg.Images))
				bgavg = bg.Image()
				analyzer.Background = bgavg
				p.Response.BackgroundImg = dataImg(bgavg, "")
			}
		}
		if p.Step == 5 && analysis.NeedsMore() {
			log.Printf("saving frame %d for analysis later", frame)
			analysis.images = append(analysis.images, rgbImg)
		}
		// set every frame, so this ends w/ the last value

		if p.Step == 6 && bgavg != nil {
			// process pending frames
			positions := analyzer.Positions(rgbImg)
			if len(positions) > 0 {
				p.Response.FramePositions = append(p.Response.FramePositions, FramePosition{frame, pkt.Time, positions})
			}
		}
		if frame == 500 {
			break
		}
	}

	if p.Step == 5 && bgavg != nil {
		analysis.Calculate(bgavg, p.Blur, p.ContiguousPixels, p.MinMass, p.Tolerance)
		p.Response.FrameAnalysis = append(p.Response.FrameAnalysis, *analysis)
	}

	return nil
}
