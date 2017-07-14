package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
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
	http.HandleFunc("/vehicle_positions.csv", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		p := NewProject(*fileName)
		p.Load(req)

	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL)
		if req.URL.Path != "/" {
			http.Error(w, "NOT_FOUND", 404)
			return
		}
		req.ParseForm()
		p := NewProject(*fileName)
		p.Load(req)

		p.Response, p.Err = p.Run()
		if p.Err != nil {
			log.Printf("%s", p.Err)
		}

		err := Template.ExecuteTemplate(w, "webpage", p)
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
