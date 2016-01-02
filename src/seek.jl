import VideoIO

# The VideoIO library is really great, but it's missing a random access seeking API.
# This should eventually be pushed upstream (https://github.com/kmsquire/VideoIO.jl/issues/30)
function Base.seek(s::VideoIO.VideoReader, time, video_stream=1)
    pCodecContext = s.pVideoCodecContext
    seek(s.avin, time, video_stream)
    VideoIO.avcodec_flush_buffers(pCodecContext)
    s
end

function Base.seek(avin::VideoIO.AVInput, time, video_stream = 1)
    # AVFormatContext
    fc = avin.apFormatContext[1]

    stream_info = avin.video_info[video_stream]

    s = stream_info.stream
    c = stream_info.codec_ctx
    base_per_frame = (c.time_base.num * c.ticks_per_frame * s.time_base.den / c.time_base.den * s.time_base.num)
    avg_frame_rate = s.avg_frame_rate.num / s.avg_frame_rate.den
    pos = floor(Int, time * base_per_frame * avg_frame_rate)
    println("seeking to $(time) sec @ position $pos (frame rate $avg_frame_rate/sec)")

    # Seek
    # pos (aka Timestamp) is in AVStream.time_base units or, if no stream is specified, in AV_TIME_BASE units.
    ret = VideoIO.av_seek_frame(fc, stream_info.stream_index0, pos, VideoIO.AVSEEK_FLAG_ANY)
    
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