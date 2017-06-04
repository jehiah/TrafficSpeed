package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func (p *Project) SetStep() {
	if p.Step != 0 {
		return
	}
	switch {
	case p.PreCrop == nil:
		p.Step = 1
	case p.Rotate == 0:
		p.Step = 2
	case p.PostCrop == nil || p.PostCrop.IsZero():
		p.Step = 3
	case len(p.Masks) == 0:
		p.Step = 4
	default:
		p.Step = 5
	}
}

func OpenInBrowser(l net.Listener) error {
	u := &url.URL{Scheme: "http", Host: l.Addr().String()}
	err := exec.Command("/usr/bin/open", u.String()).Run()
	return err
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	fileName := flag.String("file", "", "")
	httpAddress := flag.String("http-address", ":53001", "http address")
	skipBrowser := flag.Bool("skip-browser", false, "skip opening browser")
	flag.Parse()

	if *fileName == "" {
		log.Fatalf("-file required")
	}

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../data/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL)
		if req.URL.Path != "/" {
			http.Error(w, "NOT_FOUND", 404)
			return
		}
		req.ParseForm()
		p := NewProject(*fileName)

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
		p.Blur = geti64("blur", 3)
		p.MinMass = geti64("min_mass", 100)
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

		err := p.Run()
		if err != nil {
			log.Printf("%s", err)
			p.Err = err
		}

		err = Template.ExecuteTemplate(w, "webpage", p)
		if err != nil {
			log.Printf("%s", err)
		}
	})

	log.Printf("listening on %s", *httpAddress)
	listener, err := net.Listen("tcp", *httpAddress)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("Running server at %s", listener.Addr())
	if !*skipBrowser {
		go func() {
			time.Sleep(200 * time.Millisecond)
			err := OpenInBrowser(listener)
			if err != nil {
				log.Println(err)
			}
		}()
	}
	err = http.Serve(listener, nil)
	log.Printf("%s", err)
}
