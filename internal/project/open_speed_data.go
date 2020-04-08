package project

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

// exposes a UI for configuration and exits on save or cancel
func ConfigureUI(p *Project, httpAddress string) {
	s := &http.Server{Handler: http.DefaultServeMux}
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir(p.Dir))))
	http.HandleFunc("/exit", func(w http.ResponseWriter, req *http.Request) {
		s.Shutdown(context.TODO())
	})
	http.HandleFunc("/save", func(w http.ResponseWriter, req *http.Request) {
		settingsname := filepath.Join(p.Dir, "project.json")
		f, err := os.Create(settingsname)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer f.Close()
		err = json.NewEncoder(f).Encode(p)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), 500)
			return
		}
		s.Shutdown(context.TODO())
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL)
		if req.URL.Path != "/" {
			http.Error(w, "NOT_FOUND", 404)
			return
		}
		req.ParseForm()
		p.Load(req)
		p.Response, p.Err = p.Run()
		if p.Err != nil {
			log.Print(p.Err)
		}

		err := Template.ExecuteTemplate(w, "webpage", p)
		if err != nil {
			log.Print(err)
		}
	})

	log.Printf("listening on %s", httpAddress)
	listener, err := net.Listen("tcp", httpAddress)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("Running server at %s", listener.Addr())
	go func() {
		time.Sleep(200 * time.Millisecond)
		err := OpenInBrowser(listener)
		if err != nil {
			log.Println(err)
		}
	}()
	err = s.Serve(listener)
	log.Print(err)
}
