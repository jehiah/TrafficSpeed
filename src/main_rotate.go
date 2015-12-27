package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
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
	<form method="GET" action=".">
	{{ if .Error }}
		<div class="alert alert-danger" role="alert">{{.Error}}</div>
	{{ end }}

	
	<h2>Step 1: Select File</h2>
	<div class="form-group">
		<label>Filename: <input type="text" name="filename" id="filename" class="form-control" placeholder="filename.mp4" value="{{.Filename}}"></label>
	</div>

	{{ if .Filename }}
		{{ if .Rotate }}
			<h2>Step 2: Rotation</h2>
			<div class="form-group">
				<label>Ration Angle: <input name="rotate" id="rotate" type="text" value="{{.Rotate}}" /> <em>(radians)</em></label>
			</div>
			
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
	
			<img src="/data/rotate-cropped.png" id="getpoint">

		{{else}}
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
	
			<img src="/data/rotate-cropped.png" id="getpoint">
		{{ end }}
	{{ else }}
		<button type="submit" class="btn btn-primary">Continue</button>
	{{ end }}
	
	</form>
</div></div></div>
<script type="text/javascript">
function getpoint(event) {
	var i = document.getElementById("getpoint")
	var pos_x = event.offsetX ? event.offsetX : event.pageX - i.offsetLeft;
	var pos_y = event.offsetY ? event.offsetY : event.pageY - i.offsetTop;
	
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
	t.value = pos_x + "," + pos_y
	alert(title + " is " + t.value)
}
document.getElementById("getpoint").addEventListener("click", getpoint, true)

</script>
</body>
</html>`

type project struct {
	Error    error
	Filename string
	Rotate   float64
	BBox  BBox
}

type BBox struct {
	A Point
	B Point
}
func (b BBox) Range() []string {
	if b.A == b.B && b.A.X == 0 && b.A.Y == 0 {
		return nil
	}
	s := func(n float64) string {
		return fmt.Sprintf("%d", n)
	}
	return []string{
		"-x", s( math.Min(b.A.X, b.B.X)),
		"-X", s(math.Max(b.A.X, b.B.X)),
		"-y", s(math.Min(b.A.Y, b.B.Y)),
		"-Y", s(math.Max(b.A.Y, b.B.Y)),
	}
}

type Point struct {
	X float64
	Y float64
}

func ParsePoint(s string) Point {
	if !strings.Contains(s, ",") {
		return Point{}
	}
	c := strings.SplitN(s, ",", 2)
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

func (p *project) Run() {
	if p.Filename == "" {
		return
	}
	_, err := os.Stat(p.Filename)
	if err != nil {
		p.Error = err
		return
	}
	args := []string{"main_rotate.jl", "--file", p.Filename, "--output", "../data/rotate-cropped.png"}
	if p.Rotate != 0 {
		args = append(args, "--rotate", fmt.Sprintf("%0.5f", p.Rotate))
	}
	args = append(args, p.BBox.Range()...)
	
	s := time.Now()
	log.Printf("julia %s", strings.Join(args, " "))
	c := exec.Command("julia", args...)
	p.Error = c.Run()
	log.Printf("took %s", time.Since(s))
}

func main() {
	flag.Parse()
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../data/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		t := template.Must(template.New("webpage").Parse(tpl))

		req.ParseForm()
		p := &project{
			Filename: req.Form.Get("filename"),
		}

		if f, err := strconv.ParseFloat(req.Form.Get("rotate"), 64); err == nil {
			p.Rotate = f
		}

		p1, p2 := req.Form.Get("point1"), req.Form.Get("point2")
		if p.Rotate == 0 && p1 != "" && p2 != "" {
			p.Rotate = Radians(ParsePoint(p1), ParsePoint(p2))
			log.Printf("calculated rotation radians %v from a:%v b:%v", p.Rotate, p1, p2)
		}

		p.Run()

		err := t.ExecuteTemplate(w, "webpage", p)
		if err != nil {
			log.Printf("%s", err)
		}
	})

	log.Printf("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
