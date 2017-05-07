package main

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image"
	"image/png"

	"github.com/nfnt/resize"
)

func dataImg(img image.Image) (template.URL, error) {
	out := new(bytes.Buffer)

	// We now encode the image we created to the buffer
	err := png.Encode(out, img)
	if err != nil {
		return "", err
	}
	return dataImgFromBytes(out.Bytes()), nil
}

// returns a data:image/png
func dataImgFromBytes(b []byte) template.URL {
	base64Img := base64.StdEncoding.EncodeToString(b)
	return template.URL("data:image/png;base64," + base64Img)
}

func dataImgWithSize(img image.Image, width, height uint) (template.URL, error) {
	overview := resize.Thumbnail(width, height, img, resize.NearestNeighbor)
	return dataImg(overview)
}