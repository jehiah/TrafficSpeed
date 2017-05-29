package main

import (
	"html/template"
)

var Template *template.Template

func init() {
	Template = template.Must(template.New("webpage").Funcs(template.FuncMap{
		"DataURI": dataImg,
	}).Parse(tpl))
}

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


	<h2>Step 1: Video File</h2>
	<p><code>{{.Filename}}</code> 
		Frames: <code>{{.Frames}}</code> 
		Duration: <code>{{.Duration}}</code>
		Resolution: <code>{{.VideoResolution}}</code>
	</p>
	<div><img src="{{.Response.OverviewImg}}" class="img-responsive"></div>
	<input type="hidden" name="filename" value="{{.Filename}}" />
	
	
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

		<div>
		<button type="submit" class="btn" name="next" value="4">Check</button>
		<button type="submit" class="btn btn-primary" name="next" value="5">Continue</button>
		</div>
		
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
			<span class="help-block">The required difference from the background. Valid range: <kbd>0</kbd> to <kbd>255</kbd>. </span>
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
			<img src="{{.Base}}" class="img-responsive" alt="base">
			<img src="{{.BaseGif}}" class="img-responsive" alt="base-gif">

			<p>Active Image:</p>
			<img src="{{.Highlight}}" class="img-responsive" alt="highlight">
			<img src="{{.HighlightGif}}" class="img-responsive" alt="highlight-gif">

			<p>Detected Areas: (after masking)</p>
			<img src="{{.Colored}}" class="img-responsive" alt="colored">
			<img src="{{.ColoredGif}}" class="img-responsive" alt="colored-gif">

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
	
	{{ range .Response.DebugImages }}
		<img src="{{.}}" style="width: 50%; height: 50%;">
	{{ end }}
	
	</form>
</div></div>
<div class="row">&nbsp;</div>
</div>
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
