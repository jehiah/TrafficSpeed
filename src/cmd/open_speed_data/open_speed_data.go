package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	// "os/exec"
	"bytes"
	"strconv"
	"strings"
	"time"
)

const tpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" rel="stylesheet" integrity="sha256-7s5uDGW3AHqw6xtJmNNtr+OBRJUlgkNJEo78P4b0yRw= sha512-nNo+yCHEyn0smMxSswnf/OnX6/KwJuZTlNZBjauKhTK0c+zT+q5JOCx0UFhXQ6rJR9jg6Es8gPuD2uZcYDLqSw==" crossorigin="anonymous">
</head>
<body>
<div class="container"><div class="row"><div class="col-xs-12">
	<h1>Open Speed Data Analysis</h1>
	<form method="GET" action=".">
	
	{{ if .Err }}
		<div class="alert alert-danger" role="alert">{{.Err}}</div>
	{{ end }}
	{{ if .Response.Err }}
		<div class="alert alert-danger" role="alert">{{.Response.Err}}</div>
	{{ end }}

	{{ if eq .Step 1}}
		<h2>Step 1: Select Video File</h2>
		<div class="form-group">
			<label>Filename: <input type="text" name="filename" id="filename" class="form-control" placeholder="filename.mp4" value="{{.Filename}}"></label>
		</div>
	{{ else }}
		<h2>Step 1: Video File</h2>
		<p><code>{{.Filename}}</code> 
			Frames: <code>{{.Response.Frames}}</code> 
			Duration: <code>{{.Response.Duration | printf "%0.1f"}} seconds</code>
			Resolution: <code>{{.Response.VideoResolution}}</code>
		</p>
		<div><img src="{{.Response.OverviewImg}}" class="img-responsive"></div>
		<input type="hidden" name="filename" value="{{.Filename}}" />
	{{ end }}
	
	
	{{ if eq .Step 2 }}
		<h2>Step 2: Rotation Detection</h2>
		<p>Automatic rotation detection works by picking two points that align with the asis of vehicle movement</p>
		<p>Instructions: Click on the image below to select two points on the axis of movement. 
		Typically this will be a lane marking in the middle of the street at either end of the visible range. 
		After selecting "point1" and "point2" select the "Continue" button.</h2>
		<p>To skip rotation enter 0x0 as both points.</p>

		<div class="form-group">
			<label>Point 1: <input name="point1" id="point1" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Point 2: <input name="point2" id="point2" type="text" /></label>
		</div>
		<button type="submit" class="btn btn-primary">Continue</button>

		<img src="{{.Response.Step2Img}}" id="getpoint">
	{{ else if gt .Step 2 }}
		<h2>Step 2: Rotation</h2>
		<p>Rotation Angle <code>{{.Rotate}} radians</code></p>
		<div><img src="{{.Response.Step3Img}}" style="width: 25%; height: 25%;"></div>
		<input type="hidden" name="rotate" value="{{.Rotate | printf "%0.5f"}}" />
	{{ else }}
		<input type="hidden" name="rotate" value="{{.Rotate | printf "%0.5f"}}" />
	{{ end }}


	{{ if eq .Step 3 }}
		<h2>Step 3: Crop</h2>
		<p>Instructions: Click on the image below to select the upper left and lower right corner of the frame 
		to perform speed analysis on.
		After selecting "point1" and "point2" select the "Continue" button.
		</p>

		<div class="form-group">
			<label>Point 1: <input name="point1" id="point1" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Point 2: <input name="point2" id="point2" type="text" /></label>
		</div>
		<button type="submit" class="btn btn-primary">Continue</button>

		<img src="{{.Response.Step3Img}}" id="getpoint">
	{{ else if gt .Step 3 }}
		<h2>Step 3: Crop</h2>
		<p>Selected Range <code>{{.BBox}}</code>
		   Cropped Resolution: <code>{{.Response.CroppedResolution}}</code></p>
		<div><img src="{{.Response.Step4Img}}" style="width: 40%; height: 40%;"></div>
		<input type="hidden" name="bbox" value="{{.BBox}}" />
	{{ else }}
		<input type="hidden" name="bbox" value="{{.BBox}}" />
	{{ end }}

	{{ if eq .Step 4 }}
		<h2>Step 4: Mask Regions</h2>
		<p>Masking allows the detection of vehicles in different lanes to avoid bleeding into each other, 
			and eliminates irrelevant parts of the image (like sidewalks or parked cars).
			Depending on the visual perspective the masked rows should be closer to wheel position to account for 
			tall vehicles in the lane.
		</p>
		<p>Instructions: Note the X and Y from the image, and enter masks as a row range <kbd>row:row</kbd> 
			or a bounding box pair of coordinates <kbd>10x20 20x30</kbd>. To continue without masks enter a mask of <kbd>-</kbd>.</p>

		<div class="form-group">
			<label>Mask: <input name="mask" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Mask: <input name="mask" type="text" /></label>
		</div>                             
		<div class="form-group">
			<label>Mask: <input name="mask" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Mask: <input name="mask" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Mask: <input name="mask" type="text" /></label>
		</div>

		<div><button type="submit" class="btn btn-primary">Continue</button></div>
		
		<p>Mouse Position: <span id="mouse_position" style="font-weight:bold;size:14pt;"></span> <span id="mouse_click" style="font-weight:bold;size:14pt;"></span></p>

		<img src="{{.Response.Step4Img}}" id="mousemove">
	{{ else if gt .Step 4 }}
		{{ if .Masks }}
			<h2>Step 4: Mask Regions</h2>
			<p>Masked regions: {{range .Masks }}<code>{{.}}</code> {{end}}</p>
			<div><img src="{{.Response.Step4MaskImg}}" style="width: 40%; height: 40%;"></div>
			{{ range .Masks }}
				<input type="hidden" name="mask" value="{{.}}" />
			{{ end }}
		{{ end }}
	{{ else }}
		{{ range .Masks }}
			<input type="hidden" name="mask" value="{{.}}" />
		{{ end }}
	{{ end }}
	
	{{ if eq .Step 5 }}
		<h2>Step 5: Object Detection</h2>
		<p>The tunables below adjust what is detected as "active" in an image, and what is treated as a vehicle.</p>
		
		<div class="form-group">
			<label>Tolerance: <input name="tolerance" id="tolerance" type="text" value="{{.Tolerance}}" /></label>
			<span class="help-block">The required difference from the background.</span>
		</div>
		
		<div class="form-group">
			<label>Blur (pixels): <input name="blur" id="blur" type="text" value="{{.Blur}}" /></label>
			<span class="help-block">Bluring helps define features better and make a single blob for better detection.</span>
		</div>
		<div class="form-group">
			<label>Min Mass: <input name="min_mass" id="min_mass" type="text" value="{{.MinMass}}" /></label>
			<span class="help-block">Filters out small areas that are detected in the image (such as pedestrians).</span>
		</div>
		<button type="submit" class="btn" name="next" value="5">Check</button>
		<button type="submit" class="btn btn-primary" name="next" value="6">Continue</button>

		<p>Background Image:</p>
		<img src="{{.Response.BackgroundImg}}" style="width: 50%; height: 50%;">

		<div class="row">
		{{ range .Response.FrameAnalysis }}
		<div class="col-xs-12 col-md-8 col-lg-6">
			<h4>Time index <code>{{.Timestamp}} seconds</code></h4>
			<p>Frame: (before masking)</p>
			<img src="{{.Base}}" class="img-responsive">
			<p>Active Image: (before masking)</p>
			<img src="{{.HighlightGif}}" class="img-responsive">
			<p>Detected Areas: (after masking)</p>
			<img src="{{.ColoredGif}}" class="img-responsive">
			{{ if .Positions }}
				<table class="table table-striped">
				<thead>
				<tr>
					<th></th><th>Mass</th><th>Position</th><th>Size</th>
				</tr>
				</thead>
				<tbody>
				{{ range $i, $p := .Positions }}
				<tr>
					<th>{{$i}}</th>
					<td>{{$p.Mass }} pixels</td>
					<td>{{$p.X | printf "%0.f"}}x{{$p.Y | printf "%0.f"}}</td>
					<td>{{$p.Size}}</td>
				</tr>
				{{ end }}
				</tbody>
				</table>
			{{ end }}
		</div>
		{{ end }}
		</div>
	{{ else if gt .Step 5 }}
		<h2>Step 5: Object Detection</h2>
		<p>Tolerance: <code>{{.Tolerance}}</code></p>
		<p>Blur: <code>{{.Blur}}</code></p>
		<p>Min Mass: <code>{{.MinMass}}</code></p>

		<p>Background Image:</p>
		<img src="{{.Response.BackgroundImg}}" style="width: 50%; height: 50%;">

		<input type="hidden" name="tolerance" value="{{.Tolerance}}" />
		<input type="hidden" name="blur" value="{{.Blur}}" />
		<input type="hidden" name="min_mass" value="{{.MinMass}}" />
	{{ else }}
		<input type="hidden" name="tolerance" value="{{.Tolerance}}" />
		<input type="hidden" name="blur" value="{{.Blur}}" />
		<input type="hidden" name="min_mass" value="{{.MinMass}}" />
	{{ end }}
	
	{{ if eq .Step 6 }}
		<h2>Step 6: Speed Calibration</h2>
		
		<p>Calibrations: {{range .Calibrations }}<code>{{.Pretty}}</code><br/>{{end}}</p>

		{{ range .Calibrations }}
			<input type="hidden" name="calibration" value="{{.}}" />
		{{ end }}
		
		<div class="form-group">
			<label>Seek (seconds): <input name="seek" id="seek" type="text" value="{{.Seek}}" /></label>
			<button type="submit" class="btn btn-primary" name="next" value="6">Seek</button>
		</div>

		{{ if .Seek }}
			<div class="form-group">
				<label>Point 1: <input name="point1" id="point1" type="text" /></label>
			</div>
			<div class="form-group">
				<label>Point 2: <input name="point2" id="point2" type="text" /></label>
			</div>
			<div class="form-group">
				<label>Distance (inches): <input name="inches" id="inches" type="text" /></label>
				<span class="help-block">NV200 wheelbase is 115.2" </span>
			</div>
			<button type="submit" class="btn btn-primary" name="next" value="6">Record Calibration</button>
		{{ end }}
		<button type="submit" class="btn btn-primary" name="next" value="7">Done</button>
		
		<img src="{{.Response.Step6Img}}" id="getpoint">
	{{ else if gt .Step 6 }}
		{{ range .Calibrations }}
			<input type="hidden" name="calibration" value="{{.}}" />
		{{ end }}
	{{ end }}
	
	</form>
</div></div></div>
<script type="text/javascript">
function pos(el, event) {
	var pos_x = event.offsetX ? event.offsetX : event.pageX - el.offsetLeft;
	var pos_y = event.offsetY ? event.offsetY : event.pageY - el.offsetTop;
	return pos_x + "x" + pos_y
}

function getpoint(event) {
	var i = document.getElementById("getpoint")
	
	var title = "Point 1"
	var t = document.getElementById("point1")
	if (t.value != "") {
		title = "Point 2"
		t = document.getElementById("point2")
		if (t.value != "") {
			t.value = ""
			title = "Point 1"
			t = document.getElementById("point1")
		}
	}
	t.value = pos(i, event)
	alert(title + " is " + t.value)
}
function mousemove(event) {
	var i = document.getElementById("mousemove")
	document.getElementById("mouse_position").innerHTML = pos(i, event)
}
function mouseclick(event) {
	var i = document.getElementById("mousemove")
	document.getElementById("mouse_click").innerHTML = "last click: " + pos(i, event)
}
function on(pattern, event, f) {
	var el = document.getElementById(pattern)
	if (el == null) {
		return
	}
	el.addEventListener(event, f, true)
}

on("getpoint", "click", getpoint)
on("mousemove", "click", mouseclick)
on("mousemove", "mousemove", mousemove)
on("mousemove", "mouseout", function(){
	document.getElementById("mouse_position").innerHTML = ""
})

</script>
</body>
</html>`

type Project struct {
	Err          error          `json:"error,omitempty"`
	Filename     string         `json:"filename"`
	Rotate       float64        `json:"rotate,omitempty"`
	BBox         *BBox          `json:"bbox,omitempty"`
	Masks        []Mask         `json:"masks,omitempty"`
	Tolerance    float64        `json:"tolerance"`
	Blur         int64          `json:"blur"`
	MinMass      int64          `json:"min_mass"`
	Seek         float64        `json:"seek"`
	Calibrations []*Calibration `json:"calibrations"`

	Step     int      `json:"step"`
	Response Response `json:"response,omitempty"`
}

type Calibration struct {
	Seek   float64 `json:"seek"`
	A      Point   `json:"a"`
	B      Point   `json:"b"`
	Inches float64 `json:"inches"`
}

func (c *Calibration) String() string {
	return fmt.Sprintf("%0.4f %s %s %0.4f", c.Seek, c.A, c.B, c.Inches)
}
func (c *Calibration) Pretty() string {
	return fmt.Sprintf("Seek:%0.4fsec Points{%s %s} Inches:%0.4f", c.Seek, c.A, c.B, c.Inches)
}

func ParseCalibration(s string) (c *Calibration) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") || !strings.Contains(s, " ") {
		return nil
	}
	chunks := strings.SplitN(s, " ", 4)

	c = &Calibration{}
	var err error
	c.Seek, err = strconv.ParseFloat(chunks[0], 64)
	if err != nil {
		return nil
	}
	c.A = ParsePoint(chunks[1])
	c.B = ParsePoint(chunks[2])
	c.Inches, err = strconv.ParseFloat(chunks[3], 64)
	if err != nil {
		return nil
	}
	return
}

type Response struct {
	Err               string          `json:"err,omitempty"`
	Frames            int64           `json:"frames,omitempty"`
	Duration          float64         `json:"duration_seconds,omitempty"`
	VideoResolution   string          `json:"video_resolution,omitempty"`
	CroppedResolution string          `json:"cropped_resolution,omitempty"`
	OverviewGif       template.URL    `json:"overview_gif,omitempty"`
	OverviewImg       template.URL    `json:"overview_img,omitempty"`
	Step3Img          template.URL    `json:"step_3_img,omitempty"`
	Step4Img          template.URL    `json:"step_4_img,omitempty"`
	Step4MaskImg      template.URL    `json:"step_4_mask_img,omitempty"`
	BackgroundImg     template.URL    `json:"background_img,omitempty"`
	FrameAnalysis     []FrameAnalysis `json:"frame_analysis,omitempty"`
	Step6Img          template.URL    `json:"step_6_img,omitempty"`
}

type FrameAnalysis struct {
	Timestamp    float64      `json:"ts"`
	Base         template.URL `json:"base,omitempty"`
	BaseGif      template.URL `json:"base_gif,omitempty"`
	Highlight    template.URL `json:"highlight,omitempty"`
	HighlightGif template.URL `json:"highlight_gif,omitempty"`
	Colored      template.URL `json:"colored,omitempty"`
	ColoredGif   template.URL `json:"colored_gif,omitempty"`
	Positions    []Position   `json:"positions,omitempty"`
}

// Position matches Position in position.jl
type Position struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Mass  int     `json:"mass"`
	XSpan []int   `json:"xspan"` // [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
	YSpan []int   `json:"yspan"`
}

func (p Position) Span() string {
	mm := func(d []int) (min int, max int) {
		for i, n := range d {
			if n < min || i == 0 {
				min = n
			}
			if n > max || i == 0 {
				max = n
			}
		}
		return
	}
	xmin, xmax := mm(p.XSpan)
	ymin, ymax := mm(p.YSpan)
	return fmt.Sprintf("x:%d-%d y:%d-%d", xmin, xmax, ymin, ymax)
}
func (p Position) Size() string {
	return fmt.Sprintf("%dx%d", len(p.XSpan), len(p.YSpan))
}

func (p *Project) SetStep() {
	if p.Step != 0 {
		return
	}
	switch {
	case p.Filename == "":
		p.Step = 1
	case p.Rotate == 0:
		p.Step = 2
	case p.BBox == nil || p.BBox.IsZero():
		p.Step = 3
	case len(p.Masks) == 0:
		p.Step = 4
	default:
		p.Step = 5
	}
}

type Mask struct {
	Start    int64 `json:"start,omitempty"`
	End      int64 `json:"end,omitempty"`
	BBox     *BBox `json:"bbox,omitempty"`
	NullMask bool  `json:"null_mask,omitempty"`
}

func (m Mask) String() string {
	if m.NullMask {
		return "-"
	}
	if m.Start != 0 && m.End != 0 {
		return fmt.Sprintf("%d:%d", m.Start, m.End)
	}
	return m.BBox.String()
}

func ParseMask(s string) (m Mask, ok bool) {
	s = strings.TrimSpace(s)
	switch {
	case s == "-":
		m.NullMask = true
		ok = true
	case strings.Count(s, ":") == 1:
		c := strings.SplitN(s, ":", 2)
		x, _ := strconv.Atoi(c[0])
		y, _ := strconv.Atoi(c[1])
		m.Start = int64(math.Min(float64(x), float64(y)))
		m.End = int64(math.Max(float64(x), float64(y)))
		ok = true
	case strings.Count(s, "x") == 2 && strings.Count(s, " ") == 1:
		m.BBox = ParseBBox(s)
		ok = !m.BBox.IsZero()
	}
	return
}

func (p *Project) Run(backend string) error {
	p.SetStep()

	if p.Filename == "" {
		return nil
	}
	_, err := os.Stat(p.Filename)
	if err != nil {
		return err
	}

	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	s := time.Now()
	u := backend + "/api/"
	resp, err := http.Post(u, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("Got %d from %q took %s with %d bytes", resp.StatusCode, u, time.Since(s), len(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("got status code %d from %s", resp.StatusCode, u)
	}
	err = json.Unmarshal(body, &p.Response)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	backend := flag.String("backend", "http://127.0.0.1:8000", "base path to backend processing service")
	flag.Parse()

	t := template.Must(template.New("webpage").Parse(tpl))

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../data/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		req.ParseForm()
		p := &Project{
			Filename: req.Form.Get("filename"),
		}

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

		p.Rotate = getf64("rotate", 0)
		p.Tolerance = getf64("tolerance", 0.06)
		p.Blur = geti64("blur", 3)
		p.MinMass = geti64("min_mass", 100)
		p.Seek = getf64("seek", 0)
		p.BBox = ParseBBox(req.Form.Get("bbox"))
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
			case p.Rotate == 0:
				p.Rotate = Radians(ParsePoint(p1), ParsePoint(p2))
				log.Printf("calculated rotation radians %v from a:%v b:%v", p.Rotate, p1, p2)
			case p.Step == 6:
				p.Calibrations = append(p.Calibrations, &Calibration{
					Seek:   p.Seek,
					A:      ParsePoint(p1),
					B:      ParsePoint(p2),
					Inches: getf64("inches", 0),
				})
				p.Seek = 0
			default:
				p.BBox = &BBox{ParsePoint(p1), ParsePoint(p2)}
				log.Printf("Bounding Box %#v", p.BBox)
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

		err := p.Run(*backend)
		if err != nil {
			log.Printf("%s", err)
			p.Err = err
		}

		if p.BBox == nil {
			p.BBox = &BBox{} // make template easier
		}

		err = t.ExecuteTemplate(w, "webpage", p)
		if err != nil {
			log.Printf("%s", err)
		}
	})

	// TODO: goroutine exec of `julia main.jl`
	// c := exec.Command("julia", "main.jl")
	// err := c.Run()

	log.Printf("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
