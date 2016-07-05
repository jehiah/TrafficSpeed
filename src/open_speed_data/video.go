package main

import (
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"bytes"
	"image/png"

	"github.com/3d0c/gmf"
)

type Video struct {
	inputCtx *gmf.FmtCtx
	srcVideoStream *gmf.Stream
	Filename string
	Duration float64
	Frames int64
	converter *Converter
}

func Open(srcFileName string) (*Video, error) {
	v = &Video{
		inputCtx: gmf.NewInputCtx(srcFileName),
	}
	var err error
	v.srcVideoStream, err = v.inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		v.inputCtx.CloseInputAndRelease()
		return nil, err
	}
	v.Converter, err = Converter.Open(v.inputCtx)
	if err != nil {
		v.inputCtx.CloseInputAndRelease()
		return nil, err
	}
	return v, nil
}

func (v *Video) Close() {
	v.inputCtx.CloseInputAndRelease()
}
func (v *Video) Frame(n int) (*Image, error) {
	// AVSEEK_FLAG_FRAME
	// ist, err := this.GetStream(streamIndex)
	// if err := this.SeekFile(ist, frameTs, frameTs, C.AVSEEK_FLAG_FRAME); err != nil {
	// ist.CodecCtx().FlushBuffers()
	for {
		packet := v.inputCtxt.GetNextPacket()
		if packet.StreamIndex() != v.srcVideoStream.Index() {
			continue
		}
		if p != nil {
			return nil, errors.New("nil packet")
		}
	}
	return v.Converter.FrameToImage()
}
type Converter struct {
	cc *gmf.CodecCtx
	swsCtx *gmf.SwsCtx
	destFrame *gmf.Frame
}
func (c *Converter) FrameToImage(srcFrame *gmf.Frame) (*Image, error) {
	defer gmf.Release(srcFrame)
	swsCtx.Scale(srcFrame, c.dstFrame)

	p, ready, err := dstFrame.EncodeNewPacket(cc)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, errors.New("not ready")
	}
	return png.Decode(bytes.NewReader(p.Data()))
}

func (c *Converter) Close() {
	gmf.Release(c.cc)
	gmf.Release(c.swsCtx)
	gmf.Release(c.destFrame)
}

func NewConverter(srcCtx *gmf.CodecCtx) (*Converter, error) {
	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_PNG)
	if err != nil {
		return nil, err
	}

	c = &Converter{
		cc: gmf.NewCodecCtx(codec),
	}

	w, h := srcCtx.Width(), srcCtx.Height()

	c.cc.SetPixFmt(gmf.AV_PIX_FMT_RGB24).SetWidth(w).SetHeight(h)

	if codec.IsExperimental() {
		c.cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		return nil, err
	}

	c.swsCtx = gmf.NewSwsCtx(srcCtx, cc, gmf.SWS_BICUBIC)

	// convert to RGB, optionally resize could be here
	c.dstFrame = gmf.NewFrame().
		SetWidth(w).
		SetHeight(h).
		SetFormat(gmf.AV_PIX_FMT_RGB24)

	if err := dstFrame.ImgAlloc(); err != nil {
		return nil, err
	}
	return c, nil
}


func main() {
	srcFileName := "tests-sample.mp4"

	os.Mkdir("./tmp", 0755)

	if len(os.Args) > 1 {
		srcFileName = os.Args[1]
	}


	
	log.Printf("TimeBase %#v", srcVideoStream.TimeBase())
	log.Printf("Duration %#v", srcVideoStream.Duration())
	log.Printf("NbFrames %#v", srcVideoStream.NbFrames())

	codec, err := FindEncoder(AV_CODEC_ID_PNG)
	if err != nil {
		fatal(err)
	}

	cc := NewCodecCtx(codec)
	defer Release(cc)

	cc.SetPixFmt(AV_PIX_FMT_RGB24).SetWidth(srcVideoStream.CodecCtx().Width()).SetHeight(srcVideoStream.CodecCtx().Height())

	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcVideoStream.CodecCtx(), cc, SWS_BICUBIC)
	defer Release(swsCtx)

	dstFrame := NewFrame().
		SetWidth(srcVideoStream.CodecCtx().Width()).
		SetHeight(srcVideoStream.CodecCtx().Height()).
		SetFormat(AV_PIX_FMT_RGB24)
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}
	
	// AVSEEK_FLAG_FRAME
	// ist, err := this.GetStream(streamIndex)
	// if err := this.SeekFile(ist, frameTs, frameTs, C.AVSEEK_FLAG_FRAME); err != nil {
	// ist.CodecCtx().FlushBuffers()

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}
		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		for frame := range packet.Frames(ist.CodecCtx()) {
			// CloneNewFrame
			swsCtx.Scale(frame, dstFrame)

			if p, ready, _ := dstFrame.EncodeNewPacket(cc); ready {
				writeFile(p.Data())
				defer Release(p)
			}
		}
		Release(packet)
	}

	Release(dstFrame)

}
