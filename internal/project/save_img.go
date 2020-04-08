package project

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func (p *Project) SaveImage(i image.Image, name string) error {
	fname := filepath.Join(p.Dir, name)
	log.Printf("creating %s", fname)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, i)
}
