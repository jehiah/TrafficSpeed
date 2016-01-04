using Images
import VideoIO
import FixedPointNumbers

include("./seek.jl")

"""
read N random frames and generate a background from the averages
"""
function avg_background(f::VideoIO.VideoReader, rrc::Function, frames::Integer)
    seekstart(f)
    bg = float32(convert(Image{ColorTypes.RGB{Float32}}, rrc(f)))

    step = duration(f)/frames
    for count=1:frames-1 # -1 so we don't overflow seek, but also because we started w/ frame 0 as base
        println("background frame $count @ $(count*step) seconds")
        seek(f, count*step)
        frame = float32(rrc(f))
        bg += frame
    end
    bg /= frames
    convert(Image{ColorTypes.RGB{Float32}}, bg)
end