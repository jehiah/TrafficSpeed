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

	{{ if .Filename}}
		<h2>Step 1: Video File</h2>
		<p><code>{{.Filename}}</code> Frames:<code>{{.Response.Frames}}</code> Duration:<code>{{.Response.Duration | printf "%0.1f"}}seconds</code></p>
		<input type="hidden" name="filename" value="{{.Filename}}" />
		<div><img src="{{.Response.Step2Img}}" style="width: 20%; height: 20%;"></div>
	{{ end }}
	
	{{ if eq .Step "step_one" }}
		<h2>Step 1: Select Video File</h2>
		<div class="form-group">
			<label>Filename: <input type="text" name="filename" id="filename" class="form-control" placeholder="filename.mp4" value="{{.Filename}}"></label>
		</div>
	{{ end }}

	{{ if .Rotate }}
		<h2>Step 2: Rotation</h2>
		<p>Rotation Angle <code>{{.Rotate}} radians</code></p>
		<input type="hidden" name="rotate" value="{{.Rotate}}" />
		<div><img src="{{.Response.Step3Img}}" style="width: 25%; height: 25%;"></div>
	{{ end }}
	
	{{ if eq .Step "step_two" }}
		<h2>Step 2: Rotation Detection</h2>
		<p>Automatic rotation detection works by picking two points that align with the asis of vehicle movement</p>
		<p>Instructions: Click on the image below to select two points on the axis of movement. 
		Typically this will be a lane marking in the middle of the street at either end of the visible range. 
		After selecting "point1" and "point2" select the "Continue" button.</h2>

		<div class="form-group">
			<label>Point 1: <input name="point1" id="point1" type="text" /></label>
		</div>
		<div class="form-group">
			<label>Point 2: <input name="point2" id="point2" type="text" /></label>
		</div>
		<button type="submit" class="btn btn-primary">Continue</button>

		<img src="{{.Response.Step2Img}}" id="getpoint">
	{{ end }}


	{{ if eq .BBox.IsZero false}}
		<h2>Step 3: Crop</h2>
		<p>Selected Range <code>{{.BBox}}</code></p>
		<input type="hidden" name="bbox" value="{{.BBox}}" />
		<div><img src="{{.Response.Step4Img}}" style="width: 25%; height: 25%;"></div>
	{{ end }}
	
	{{ if eq .Step "step_three" }}
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
	{{ end }}

	{{ if .Masks }}
		<h2>Step 4: Mask Regions</h2>
		<p>Masked regions: {{range .Masks }}<code>{{.}}</code> {{end}}</p>
	{{ end }}
	
	{{ if eq .Step "step_four" }}
		<h2>Step 4: Mask Regions</h2>
		<p>Masking allows the detection of vehicles in different lanes to avoid bleeding into each other, 
			and eliminates irrelevant parts of the image (like sidewalks or parked cars).
			Depending on the visual perspective the masked rows should be closer to wheel position to account for 
			tall vehicles in the lane.
		</p>
		<p>Instructions: Note the X and Y from the image, and enter masks as a row range <code>row:row</code> 
			or a bounding box pair of coordinates <code>10x20 20x30</code>. To continue without masks enter a mask of <code>-</code>.</p>

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
	{{ end }}
	
	{{ if eq .Step "step_five" }}
		<h2>Step 5: Object Detection</h2>
		<p>The threshold must be set for what size triggers vehicle detection.</p>
		<p>Three frames have been randomly selected as examples.</p>
		<div class="row">
		<img src="{{.Response.Step5Img}}" id="mousemove">
	{{ end }}
	
	{{ if eq .Step "step_six"}}
		<h2>Step 6: Speed Detection</h2>
		
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
	Err      error    `json:"error,omitempty"`
	Filename string   `json:"filename"`
	Rotate   float64  `json:"rotate,omitempty"`
	BBox     BBox     `json:"bbox,omitempty"`
	Masks    []Mask   `json:"masks,omitempty"`
	Step     string   `json:"step"`
	Response Response `json:"response,omitempty"`
}
type Response struct {
	Err      string       `json:"err,omitempty"`
	Frames   int64        `json:"frames,omitempty"`
	Duration float64      `json:"duration_seconds,omitempty"`
	Step2Img template.URL `json:"step_2_img,omitempty"`
	Step3Img template.URL `json:"step_3_img,omitempty"`
	Step4Img template.URL `json:"step_4_img,omitempty"`
	Step5Img template.URL `json:"step_5_img,omitempty"`
}

func (p *Project) getStep() string {
	switch {
	case p.Filename == "":
		return "step_one"
	case p.Rotate == 0:
		return "step_two"
	case p.BBox.IsZero():
		return "step_three"
	case len(p.Masks) == 0:
		return "step_four"
	default:
		return "step_five"
	}
}

type Mask struct {
	Start    int64 `json:"start,omitempty"`
	End      int64 `json:"end,omitempty"`
	BBox     BBox  `json:"bbox,omitempty"`
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
	case strings.Count(s, "x") == 2 && strings.Count(s, " ") == 2:
		m.BBox = ParseBBox(s)
		ok = !m.BBox.IsZero()
	}
	return
}

type BBox struct {
	A Point `json:"a"`
	B Point `json:"b"`
}

func ParseBBox(s string) (b BBox) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") || !strings.Contains(s, " ") {
		return
	}
	c := strings.SplitN(s, " ", 2)
	p1 := ParsePoint(c[0])
	p2 := ParsePoint(c[1])
	// for a bounding box, always top left and bottom right
	b.A.X = math.Min(p1.X, p2.X)
	b.A.Y = math.Min(p1.Y, p2.Y)
	b.B.X = math.Max(p1.X, p2.X)
	b.B.Y = math.Max(p1.Y, p2.Y)
	return
}

func (b BBox) IsZero() bool {
	if b.A.X == 0 && b.A.Y == 0 && b.B.X == 0 && b.B.Y == 0 {
		return true
	}
	return false
}
func (b BBox) String() string {
	return fmt.Sprintf("%s %s", b.A, b.B)
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p Point) String() string {
	return fmt.Sprintf("%dx%d", int64(p.X), int64(p.Y))
}

func ParsePoint(s string) (p Point) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") {
		return
	}
	c := strings.SplitN(s, "x", 2)
	x, _ := strconv.Atoi(c[0])
	y, _ := strconv.Atoi(c[1])
	return Point{float64(x), float64(y)}
}

func Radians(a, b Point) float64 {
	if a.Y == b.Y {
		return 0
	}

	adjacent := math.Max(a.X, b.X) - math.Min(a.X, b.X)
	opposite := math.Max(a.Y, b.Y) - math.Min(a.Y, b.Y)
	radians := math.Atan(adjacent / opposite)
	log.Printf("adjacent: %v opposite %v radians %v", adjacent, opposite, radians)
	if a.Y < b.Y {
		return (-1 * radians) + 1.570796
	}
	return radians
}

func (p *Project) Run(backend string) error {
	p.Step = p.getStep()

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
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../data/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		t := template.Must(template.New("webpage").Parse(tpl))

		req.ParseForm()
		p := &Project{
			Filename: req.Form.Get("filename"),
		}

		if f, err := strconv.ParseFloat(req.Form.Get("rotate"), 64); err == nil {
			p.Rotate = f
		}
		p.BBox = ParseBBox(req.Form.Get("bbox"))

		p1, p2 := req.Form.Get("point1"), req.Form.Get("point2")
		if p.Rotate == 0 && p1 != "" && p2 != "" {
			p.Rotate = Radians(ParsePoint(p1), ParsePoint(p2))
			log.Printf("calculated rotation radians %v from a:%v b:%v", p.Rotate, p1, p2)
		} else if p1 != "" && p2 != "" {
			p.BBox = BBox{ParsePoint(p1), ParsePoint(p2)}
			log.Printf("Bounding Box %#v", p.BBox)
		}

		for i, m := range req.Form["mask"] {
			if mm, ok := ParseMask(m); ok {
				p.Masks = append(p.Masks, mm)
			} else if !ok && len(strings.TrimSpace(m)) > 0 {
				p.Err = fmt.Errorf("Error Parsing Mask #%d %q", i, m)
				break
			}
		}

		err := p.Run(*backend)
		if err != nil {
			log.Printf("%s", err)
			p.Err = err
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
