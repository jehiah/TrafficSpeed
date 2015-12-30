using Images
import VideoIO
import FixedPointNumbers

include("./seek.jl")

"""
read 10 random frames and generate a background from the averages
"""
function avg_background(f::VideoIO.VideoReader, rrc::Function)
    seekstart(f)

    bg = similar(rrc(f))
    println("rrc summary $(summary(bg))")
    println("bg is $(bg.data[1:3,1:1])")
    # bg = convert(Image{ColorTypes.RGB{Float32}}, frame)

    i = 1.5
    total_pos = 0
    total_duration = duration(f)
    count = 1
    while total_pos + i < total_duration
        println("bg is $(bg.data[1:3,1:1])")
        total_pos += i
        count+=1
        println("generating background: seeking $i to $total_pos for frame $count")
        seek(f, i)
        
        bg += rrc(f)
        
        i *= 2
    end
    println("bg is $(bg.data[1:3,1:1])")
    println("bg/=count($count)")
    bg /= count
    println("bg is $(bg.data[1:3,1:1])")
    for i in eachindex(bg)
        bg[i] = clamp(bg[i],0.0,1.0)
    end
    
    bg
    # println("before convert $(summary(bg))")
    # shareproperties(frame, reinterpret(ColorTypes.RGB{Float64}, bg.data))
end