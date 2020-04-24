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
	"github.com/nareix/joy4/format/mp4"
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
	decoded  bool
	vf       *ffmpeg.VideoFrame
}

func (p *Iterator) Close() {
	if p.demuxer != nil {
		p.demuxer.Close()
		p.demuxer = nil
	}
}

func NewIterator(filename string) (iter *Iterator, err error) {
	if filename == "" {
		panic("missing filename")
	}
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
		fmt.Printf("stream[%d] = %s (video:%v)\n", i, stream.Type(), stream.Type().IsVideo())
		if stream.Type().IsVideo() {
			vstream := stream.(av.VideoCodecData)
			r := image.Rect(0, 0, vstream.Width(), vstream.Height())
			if iter.rect.Empty() {
				iter.rect = r
			} else if !iter.rect.Eq(r) {
				return nil, fmt.Errorf("video stream %d(%v) doesn't match expected %v", i, r, iter.rect)
			}
			// fmt.Printf("stream[%d] = %s\n", i, vstream.Type())
			// fmt.Printf("stream[%d] %#v\n", i, vstream)
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

func (i *Iterator) Seek(d time.Duration) error {
	dm := i.demuxer.(*avutil.HandlerDemuxer).Demuxer.(*mp4.Demuxer)
	log.Printf("should seek %s", d)
	return dm.SeekToTime(d)
}

func (i *Iterator) VideoResolution() string {
	return fmt.Sprintf("%dx%d", i.rect.Dx(), i.rect.Dy())
}

func (i *Iterator) NextWithImage() bool {
	for i.Next() {
		i.err = i.DecodeFrame()
		if i.err != nil {
			return false
		}
		if i.vf == nil {
			continue
		}
		return true
	}
	return false

}
func (i *Iterator) Next() bool {
	var err error
	var pkt av.Packet
	for {
		i.vf = nil
		i.decoded = false
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
		return true
	}
}

func (i *Iterator) Frame() int {
	return i.frame
}

func (i *Iterator) DecodeFrame() error {
	if i.decoded {
		return i.err
	}
	// decode
	decoder := i.decoders[i.packet.Idx]
	var err error
	if len(i.packet.Data) == 0 {
		log.Printf("no packet at frame %d", i.frame)
		return nil
	}
	i.vf, err = decoder.Decode(i.packet.Data)
	if i.vf == nil {
		log.Printf("no image at frame %d", i.frame)
		i.frame--
	}
	i.decoded = true
	return err
}

func (i *Iterator) Image() *image.YCbCr {
	if i.frame == -1 {
		if !i.NextWithImage() {
			panic("no image")
		}
	}
	i.err = i.DecodeFrame()
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
