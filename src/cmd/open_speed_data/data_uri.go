package main

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image"
	"image/png"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

func dataImg(img image.Image, mime string) template.URL {
	if mime == "" {
		mime = "image/webp"
	}
	out := new(bytes.Buffer)
	var err error
	switch mime {
	case "image/png":
		err = png.Encode(out, img)
	case "image/webp":
		err = webp.Encode(out, img, &webp.Options{Quality: 75})
	default:
		panic("unknown type " + mime)
	}
	// We now encode the image we created to the buffer
	if err != nil {
		panic(err.Error())
	}
	return dataImgFromBytes(mime, out.Bytes())
}

// returns a data:image/png
func dataImgFromBytes(mime string, b []byte) template.URL {
	base64Img := base64.StdEncoding.EncodeToString(b)
	return template.URL("data:" + mime + ";base64," + base64Img)
}

func dataImgWithSize(img image.Image, width, height uint, mime string) template.URL {
	overview := resize.Thumbnail(width, height, img, resize.NearestNeighbor)
	return dataImg(overview, mime)
}
