package main

// #cgo LDFLAGS: -L/usr/local/Cellar/ffmpeg/3.3/lib

import (
	"fmt"
	"log"
	"flag"
	"time"
	
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/cgo/ffmpeg"
)

func init() {
	format.RegisterAll()
}

func main() {
	name := flag.String("name", "", "filename to extact")
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	
	file, err := avutil.Open(*name)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer file.Close()

	streams, err := file.Streams()
	if err != nil {
		log.Fatalf("%s", err)
	}
	decoders := make([]*ffmpeg.VideoDecoder, len(streams))
	for i, stream := range streams {
		if stream.Type().IsAudio() {
			astream := stream.(av.AudioCodecData)
			fmt.Println(astream.Type(), astream.SampleRate(), astream.SampleFormat(), astream.ChannelLayout())
		} else if stream.Type().IsVideo() {
			vstream := stream.(av.VideoCodecData)
			fmt.Println(vstream.Type(), vstream.Width(), vstream.Height())
			fmt.Printf("%#v\n", vstream)
			decoders[i], err = ffmpeg.NewVideoDecoder(vstream)
			if err != nil {
				log.Fatalf("%s", err)
			}
		}
	}

	for i := 0; i < 20; {
		var pkt av.Packet
		var err error
		if pkt, err = file.ReadPacket(); err != nil {
			break
		}
		if !streams[pkt.Idx].Type().IsVideo() {
			continue
		}
		ms := (pkt.Time / time.Millisecond) * time.Millisecond
		fmt.Println("pkt", i, streams[pkt.Idx].Type(), ms, "len", len(pkt.Data), "keyframe", pkt.IsKeyFrame)
		decoder := decoders[pkt.Idx]
		vf, err := decoder.Decode(pkt.Data)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Printf("%#v\n", vf.Image)
		// 
		// VideoDecoder.Decode(pkt ([]bytes))
		// VideoDecoder.Image // YCbCr
		i++
	}
}
