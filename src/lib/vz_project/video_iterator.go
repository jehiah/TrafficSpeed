package project

import (
	"fmt"
	"image"
	"io"
	"log"
	"time"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
)

func init() {
	format.RegisterAll()
}

type Iterator struct {
	err      error
	demuxer  av.DemuxCloser
	decoders []*ffmpeg.VideoDecoder
	rect     image.Rectangle
	packet   av.Packet
	frame    int
	vf       *ffmpeg.VideoFrame
}

func (p *Iterator) Close() {
	if p.demuxer != nil {
		p.demuxer.Close()
		p.demuxer = nil
	}
}

func NewIterator(filename string) (iter *Iterator, err error) {
	iter = &Iterator{frame: -1}
	iter.demuxer, err = avutil.Open(filename)
	if err != nil {
		return nil, err
	}

	streams, err := iter.demuxer.Streams()
	if err != nil {
		iter.Close()
		return nil, err
	}
	iter.decoders = make([]*ffmpeg.VideoDecoder, len(streams))
	for i, stream := range streams {
		// if stream.Type().IsAudio() {
		// astream := stream.(av.AudioCodecData)
		// fmt.Println(astream.Type(), astream.SampleRate(), astream.SampleFormat(), astream.ChannelLayout())
		// } else if stream.Type().IsVideo() {
		if stream.Type().IsVideo() {
			vstream := stream.(av.VideoCodecData)
			r := image.Rect(0, 0, vstream.Width(), vstream.Height())
			if iter.rect.Empty() {
				iter.rect = r
			} else if !iter.rect.Eq(r) {
				return nil, fmt.Errorf("video stream %d(%v) doesn't match expected %v", i, r, iter.rect)
			}
			fmt.Println(vstream.Type())
			fmt.Printf("stream[%d] %#v\n", i, vstream)
			iter.decoders[i], err = ffmpeg.NewVideoDecoder(vstream)
			if err != nil {
				log.Fatalf("NewVideoDecoder error: %s", err)
			}
		}
	}
	if iter.rect.Empty() {
		return nil, fmt.Errorf("no video stream found")
	}
	return iter, nil
}

func (i *Iterator) VideoResolution() string {
	return fmt.Sprintf("%dx%d", i.rect.Dx(), i.rect.Dy())
}

func (i *Iterator) Next() bool {
	var err error
	var pkt av.Packet
	for {
		if pkt, err = i.demuxer.ReadPacket(); err != nil {
			if err == io.EOF {
				return false
			}
			i.err = err
			return false
		}
		// skip packets we don't have a decoder for
		if i.decoders[pkt.Idx] == nil {
			continue
		}
		i.packet = pkt
		i.frame++

		// decode
		decoder := i.decoders[pkt.Idx]
		i.vf, err = decoder.Decode(pkt.Data)
		if err != nil {
			i.err = err
			return false
		}
		if i.vf == nil {
			log.Printf("no video frame?")
			i.frame--
			continue
		}
		return true
	}
}

func (i *Iterator) Frame() int {
	return i.frame
}
func (i *Iterator) Image() *image.YCbCr {
	if i.vf == nil {
		return nil
	}
	return &i.vf.Image
}
func (i *Iterator) Error() error            { return i.err }
func (i *Iterator) Duration() time.Duration { return i.packet.Time }
func (i *Iterator) DurationMs() time.Duration {
	return (i.packet.Time / time.Millisecond) * time.Millisecond
}
func (i *Iterator) IsKeyFrame() bool { return i.packet.IsKeyFrame }
