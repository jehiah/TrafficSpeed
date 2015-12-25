package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
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
	<form method="POST" action=".">
		{{ if .Error }}
			<div class="alert alert-danger" role="alert">{{.Error}}</div>
		{{ end }}

		<div class="form-group">
			<label><input type="text" name="filename" id="filename" class="form-control" placeholder="filename.mp4" value="{{.Filename}}"></label>
		</div>
		{{ if .Filename }}
		<div class="form-group">
			<label><input name="rotate" id="rotate" type="range" style="width:100%;" value="{{.Rotate}}" min="-1.8" max="1.8" step=".001" onchange="document.getElementById('rotate-value').innerHTML = this.value;" /> <span id="rotate-value">{{.Rotate}}</span></label>
		</div>
		{{ end }}
		
		<img src="/data/rotate-cropped.png">
		
		<button type="submit" class="btn btn-default">Update</button>
	</form>
</div></div></div>
</body>
</html>`

type project struct {
	Error    error
	Filename string
	Rotate   float64
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
		args = append(args, "--rotate", fmt.Sprintf("%0.3f", p.Rotate), "-x", "-1000", "-X", "1000", "-y", "-1000", "-Y", "1000")
	}
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
		p.Run()

		err := t.ExecuteTemplate(w, "webpage", p)
		if err != nil {
			log.Printf("%s", err)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
