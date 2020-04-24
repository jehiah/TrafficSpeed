package project

// #cgo LDFLAGS="-L/usr/local/Cellar/ffmpeg/4.2.2_2/lib"
// #cgo CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/4.2.2_2/include"

import (
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/jehiah/TrafficSpeed/img/avgimg"
	"github.com/jehiah/TrafficSpeed/img/imgutils"
)

// roughly 7s of frames
const bgFrameCount = 15
const bgFrameSkip = 15

type Project struct {
	Filename string `json:"filename"` // Filename is not relative to Dir
	Dir      string `json:"dir,omitempty"`
	Settings `json:"settings"`

	Duration        time.Duration `json:"duration,omitempty"`
	VideoResolution string        `json:"video_resolution,omitempty"`
	Frames          int64         `json:"frames,omitempty"`

	Seek     float64  `json:"-"`
	Step     int      `json:"-"`
	Response Response `json:"-"`

	Err error `json:"-"`
}

// Settings are the user configurable options
type Settings struct {
	PreCrop          *BBox          `json:"pre_crop,omitempty"`
	Rotate           float64        `json:"rotate,omitempty"` // radians
	PostCrop         *BBox          `json:"post_crop,omitempty"`
	Masks            Masks          `json:"masks,omitempty"`
	Tolerance        uint8          `json:"tolerance,omitempty"`
	Blur             int            `json:"blur,omitempty"`
	ContiguousPixels int            `json:"contiguous_pixels,omitempty"`
	MinMass          int            `json:"min_mass,omitempty"`
	Calibrations     []*Calibration `json:"-"`
}

type Response struct {
	PreCroppedResolution string          `json:"pre_cropped_resolution,omitempty"`
	CroppedResolution    string          `json:"cropped_resolution,omitempty"`
	OverviewGif          template.URL    `json:"overview_gif,omitempty"`
	Step4Img             template.URL    `json:"step_4_img,omitempty"` // mask
	Step4MaskImg         template.URL    `json:"step_4_mask_img,omitempty"`
	BackgroundImg        template.URL    `json:"background_img,omitempty"`
	FrameAnalysis        []FrameAnalysis `json:"frame_analysis,omitempty"`
	VehiclePositions     []VehiclePosition
	Step6Img             template.URL `json:"step_6_img,omitempty"`

	DebugImages []template.URL
}

type frameImage struct {
	Frame int
	Time  time.Duration
	Image *image.RGBA
}

// NewProject starst a project for the specified video file
func NewProject(f string, iterator *Iterator) *Project {
	p := &Project{
		Filename:        f,
		VideoResolution: iterator.VideoResolution(),
	}
	return p
}

// LoadProject loads a project from a JSON file
func LoadProject(f string) (*Project, error) {
	var p Project
	r, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(r).Decode(&p)
	if err != nil {
		return nil, err
	}
	p.Dir = filepath.Dir(f)
	return &p, nil
}

func (p *Project) Iterator() (*Iterator, error) {
	videoFile := filepath.Join(p.Dir, p.Filename)
	iterator, err := NewIterator(videoFile)
	return iterator, err
}

func (p *Project) Load(req *http.Request) error {
	getf64 := func(key string, d float64) float64 {
		if v := req.Form.Get(key); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
		return d
	}
	geti64 := func(key string, d int64) int64 {
		if v := req.Form.Get(key); v != "" {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
		return d
	}
	getuint8 := func(key string, d uint8) uint8 {
		if v := req.Form.Get(key); v != "" {
			if i, err := strconv.ParseUint(v, 10, 8); err == nil {
				return uint8(i)
			}
		}
		return d
	}

	p.PreCrop = ParseBBox(req.Form.Get("pre_crop"))
	p.Rotate = getf64("rotate", 0)
	p.PostCrop = ParseBBox(req.Form.Get("post_crop"))
	p.Tolerance = getuint8("tolerance", 40)
	p.Blur = int(geti64("blur", 2))
	p.ContiguousPixels = int(geti64("contiguous_pixels", 3))
	p.MinMass = int(geti64("min_mass", 50))
	p.Seek = getf64("seek", 0)
	p.Step = int(geti64("next", 0))

	for _, s := range req.Form["calibration"] {
		c := ParseCalibration(s)
		if c != nil {
			p.Calibrations = append(p.Calibrations, c)
		} else {
			log.Printf("error parsing calibration %q", s)
		}
	}

	p1, p2 := req.Form.Get("point1"), req.Form.Get("point2")
	if p1 != "" && p2 != "" {
		switch {
		case p.Step == 2:
			p.PreCrop = &BBox{ParsePoint(p1), ParsePoint(p2)}
		case p.Rotate == 0 && p.Step == 3:
			p.Rotate = Radians(ParsePoint(p1), ParsePoint(p2))
			log.Printf("calculated rotation radians %v from a:%v b:%v", p.Rotate, p1, p2)
		case p.Step == 4:
			p.PostCrop = &BBox{ParsePoint(p1), ParsePoint(p2)}
		case p.Step == 6:
			p.Calibrations = append(p.Calibrations, &Calibration{
				Seek:   p.Seek,
				A:      ParsePoint(p1),
				B:      ParsePoint(p2),
				Inches: getf64("inches", 0),
			})
			p.Seek = 0
		default:
			log.Panicf("unknown point for step %v", p.Step)
		}
	}

	for i, m := range req.Form["mask"] {
		if mm, ok := ParseMask(m); ok {
			p.Masks = append(p.Masks, mm)
		} else if !ok && len(strings.TrimSpace(m)) > 0 {
			p.Err = fmt.Errorf("Error Parsing Mask #%d %q", i+1, m)
			break
		}
	}
	p.Masks = p.Masks.Uniq()
	return nil
}

func (p *Project) Run() (Response, error) {
	start := time.Now()
	defer func() {
		log.Printf("Run took %s", time.Since(start))
	}()

	p.SetStep()
	var results Response
	var analysis = &FrameAnalysis{}

	iterator, err := p.Iterator()
	if err != nil {
		return results, err
	}
	defer iterator.Close()

	// set overview img
	bg := &avgimg.MedianRGBA{}
	var bgavg *image.RGBA
	var framePositions []FramePosition
	var pendingAnalysis []frameImage
	analyzer := &Analyzer{
		BWCutoff:          p.Tolerance,
		BlurRadius:        p.Blur,
		ContinguousPixels: p.ContiguousPixels,
		MinMass:           p.MinMass,
	}

	setFrame := p.Frames == 0
	log.Printf("step %v", p.Step)
	for iterator.Next() {
		frame := iterator.Frame()
		// log.Printf("frame %d time %s", frame, iterator.Duration())
		if setFrame {
			p.Frames = int64(frame)
			p.Duration = iterator.Duration()
		}
		// log.Printf("loop frame:%d duration:%s", frame, iterator.Duration())
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

		var rgbImg *image.RGBA
		var img *image.RGBA
		if interested {
			log.Printf("interested in frame %d time %s", frame, iterator.Duration())
			if yi := iterator.Image(); yi == nil {
				log.Printf("no image. continuing")
				continue
			} else {
				rgbImg = imgutils.RGBA(yi)
			}
			img = rgbImg
		}

		if frame == 0 {

			if p.PreCrop != nil {
				log.Printf("PreCrop %v", p.PreCrop)
				img = img.SubImage(p.PreCrop.Rect()).(*image.RGBA)
				results.PreCroppedResolution = fmt.Sprintf("%dx%d", p.PreCrop.Dx(), p.PreCrop.Dy())
			}

			if p.Step == 2 {
				err = p.SaveImage(img, "step2_crop.png")
				if err != nil {
					return results, err
				}
			}

			if p.Step >= 3 {
				log.Printf("rotating %v", p.Rotate)
				// apply rotation

				img = transform.Rotate(img, RadiansToDegrees(p.Rotate), &transform.RotationOptions{ResizeBounds: true})
				err = p.SaveImage(img, "step3_rotate.png")
				if err != nil {
					return results, err
				}
			}
			if p.Step >= 4 {
				// rotate & crop
				log.Printf("PostCrop %v", p.PostCrop)
				if p.PostCrop != nil {
					img = img.SubImage(p.PostCrop.Rect()).(*image.RGBA)
					// img = transform.Crop(img, p.BBox.Rect())
					results.CroppedResolution = fmt.Sprintf("%dx%d", p.PostCrop.Dx(), p.PostCrop.Dy())
				} else {
					results.CroppedResolution = fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy())
				}
				results.Step4Img = dataImg(img, "image/webp")
			}
			if p.Step >= 5 {
				// mask
				p.Masks.Apply(img)
				results.Step4MaskImg = dataImg(img, "image/webp")
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
				results.BackgroundImg = dataImg(bgavg, "")
			}
		}
		if p.Step == 5 && analysis.NeedsMore() {
			log.Printf("saving frame %d for analysis later", frame)
			analysis.images = append(analysis.images, rgbImg)
		}

		// set every frame, so this ends w/ the last value
		if p.Step == 6 && bgavg == nil {
			pendingAnalysis = append(pendingAnalysis, frameImage{frame, iterator.Duration(), rgbImg})
		}

		if p.Step == 6 && bgavg != nil {
			// process pending frames
			log.Printf("extracting vehicle position from %d pending frames", len(pendingAnalysis))
			for _, pf := range pendingAnalysis {
				if pf.Frame%50 == 0 && pf.Frame > 0 {
					log.Printf("... frame %d", pf.Frame)
				}
				positions := analyzer.Positions(pf.Image)
				if len(positions) > 0 {
					framePositions = append(framePositions, FramePosition{pf.Frame, pf.Time, positions})
				}
			}
			pendingAnalysis = nil
			positions := analyzer.Positions(rgbImg)
			if len(positions) > 0 {
				framePositions = append(framePositions, FramePosition{frame, iterator.Duration(), positions})
			}
		}
		if p.Step <= 1 && frame == 0 {
			break
		}
		if p.Step == 6 && frame >= 200 {
			break
		}
	}
	if err = iterator.Error(); err != nil {
		return results, err
	}
	if p.Step >= 5 && bgavg == nil {
		if len(bg.Images) > 0 {
			log.Printf("calculating background from %d frames", len(bg.Images))
			bgavg = bg.Image()
			analyzer.Background = bgavg
			results.BackgroundImg = dataImg(bgavg, "")

			if p.Step == 6 {
				log.Printf("extracting vehicle position from %d pending frames", len(pendingAnalysis))
				for _, pf := range pendingAnalysis {
					if pf.Frame%50 == 0 && pf.Frame > 0 {
						log.Printf("... frame %d", pf.Frame)
					}
					positions := analyzer.Positions(pf.Image)
					if len(positions) > 0 {
						framePositions = append(framePositions, FramePosition{pf.Frame, pf.Time, positions})
					}
				}
				pendingAnalysis = nil
			}
		}

	}

	if p.Step == 6 {
		results.VehiclePositions = TrackVehicles(framePositions)
	}

	if p.Step == 5 && bgavg != nil {
		analysis.Calculate(bgavg, p.Blur, p.ContiguousPixels, p.MinMass, p.Tolerance)
		results.FrameAnalysis = append(results.FrameAnalysis, *analysis)
	}

	return results, nil
}
