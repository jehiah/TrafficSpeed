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

    # Get stream information
    stream_info = avin.video_info[video_stream]
    seek_stream_index = stream_info.stream_index0
    stream = stream_info.stream
    time_base = stream_info.codec_ctx.time_base
    ticks_per_frame = stream_info.codec_ctx.ticks_per_frame
    
    # time_base.den - ticks per second
    ticks_per_second = Float32(time_base.den)/ Float32(time_base.num)
    frames_per_second = ticks_per_second/Float32(ticks_per_frame)
    
    pos = floor(Int, time * (time_base.den / ticks_per_frame))
    println("seeking ahead $(time) sec by increasing position $pos (frame rate $frames_per_second/sec)")

    # Seek
    # pos (aka Timestamp) is in AVStream.time_base units or, if no stream is specified, in AV_TIME_BASE units.
    ret = VideoIO.av_seek_frame(fc, seek_stream_index, pos, VideoIO.AVSEEK_FLAG_ANY)
    
    ret < 0 && throw(ErrorException("Could not seek to position of stream"))

    return avin
end

function duration(avin::VideoIO.AVInput, video_stream = 1)
    stream_info = avin.video_info[video_stream]
    time_base = stream_info.codec_ctx.time_base

    ticks_per_second = Float32(time_base.den)/ Float32(time_base.num)
    ticks_per_frame = stream_info.codec_ctx.ticks_per_frame
    frames_per_second = ticks_per_second/Float32(ticks_per_frame)
    frame_count = stream_info.stream.nb_frames
    d = Float32(frame_count) / frames_per_second
    println("duration $d - frame_count $frame_count, fps $frames_per_second tps $ticks_per_second tpf $ticks_per_frame time_base $time_base")
    d
end
duration(s::VideoIO.VideoReader, video_stream=1) = duration(s.avin, video_stream)

# While we're at it, It's very handy to know how many frames there are:
Base.length(s::VideoIO.VideoReader) = s.avin.video_info[1].stream.nb_frames