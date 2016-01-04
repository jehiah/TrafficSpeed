import VideoIO

# The VideoIO library is really great, but it's missing a random access seeking API.
# This should eventually be pushed upstream (https://github.com/kmsquire/VideoIO.jl/issues/30)
function Base.seek(s::VideoIO.VideoReader, time, video_stream=1)
    pCodecContext = s.pVideoCodecContext
    seek(s.avin, time, video_stream)
    VideoIO.avcodec_flush_buffers(pCodecContext)
    s
end

function av_rescale_q(a, bq::VideoIO.AVRational, cq::VideoIO.AVRational)
    b = bq.num * cq.den
    c = cq.num * bq.den
    return a * b / c
end

function Base.seek(avin::VideoIO.AVInput, time, video_stream = 1)
    # AVFormatContext
    fc = avin.apFormatContext[1]
    stream_info = avin.video_info[video_stream]

    # https://www.ffmpeg.org/doxygen/2.3/group__lavu__time.html
    seek_target = time * VideoIO.AV_TIME_BASE

    # http://dranger.com/ffmpeg/functions.html#av_rescale_q
    # seek_target= av_rescale_q(seek_target, AV_TIME_BASE_Q, pFormatCtx->streams[stream_index]->time_base);
    seek_target = floor(Int, av_rescale_q(seek_target, VideoIO.AVUtil.AV_TIME_BASE_Q, stream_info.stream.time_base))
    
    # pos (aka Timestamp) is in AVStream.time_base units or, if no stream is specified, in AV_TIME_BASE units.
    # https://www.ffmpeg.org/doxygen/2.5/group__lavf__decoding.html#gaa23f7619d8d4ea0857065d9979c75ac8
    # http://dranger.com/ffmpeg/functions.html#av_seek_frame
    ret = VideoIO.av_seek_frame(fc, stream_info.stream_index0, seek_target, VideoIO.AVSEEK_FLAG_ANY)
    
    ret < 0 && throw(ErrorException("Could not seek to position of stream"))

    return avin
end

function video_summary(v::VideoIO.VideoReader, video_stream=1)
    stream_info = v.avin.video_info[video_stream]
    s = stream_info.stream
    c = stream_info.codec_ctx
    avg_frame_rate = s.avg_frame_rate.num / s.avg_frame_rate.den
    println("duration: $(duration(v)) sec, $(length(v)) frames. frame rate: $avg_frame_rate/sec")
    println("codec time_base $(c.time_base) stream time_base $(s.time_base)")
end

function duration(avin::VideoIO.AVInput, video_stream = 1)
    stream_info = avin.video_info[video_stream]
    return stream_info.stream.duration / stream_info.stream.time_base.den
end
duration(s::VideoIO.VideoReader, video_stream=1) = duration(s.avin, video_stream)

# While we're at it, It's very handy to know how many frames there are:
Base.length(s::VideoIO.VideoReader) = s.avin.video_info[1].stream.nb_frames