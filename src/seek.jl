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

    pos = Int(div(time*time_base.den, time_base.num*ticks_per_frame))
    # println("seek $pos in $video_stream time_base:$time_base ticks_per_frame:$ticks_per_frame seek_stream_index:$seek_stream_index")
    # Seek
    ret = VideoIO.av_seek_frame(fc, seek_stream_index, pos, VideoIO.AVSEEK_FLAG_ANY)

    ret < 0 && throw(ErrorException("Could not seek to start of stream"))

    return avin
end

# While we're at it, It's very handy to know how many frames there are:
Base.length(s::VideoIO.VideoReader) = s.avin.video_info[1].stream.nb_frames