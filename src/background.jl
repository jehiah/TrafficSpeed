using Images
import VideoIO
import FixedPointNumbers

include("./seek.jl")

"""
read 10 random frames and generate a background from the averages
"""
function avg_background(f::VideoIO.VideoReader, rrc::Function)
    seekstart(f)
    bg = float32(convert(Image{ColorTypes.RGB{Float32}}, rrc(f)))
    println("rrc summary $(summary(bg))")
    # bg = convert(Image{ColorTypes.RGB{Float32}}, frame)

    step = duration(f)/30
    total_pos = 0
    count = 1
    while count < 30
        total_pos += step
        count+=1
        println("generating background: seeking $step to $total_pos for frame $count")
        seek(f, total_pos)
        frame = float32(rrc(f))
        bg += frame
    end
    bg /= count
    convert(Image{ColorTypes.RGB{Float32}}, bg)
end