# https://github.com/mbauman/TrafficSpeed
# 
# Let's create a GIF to display a snippet of the raw footage. There aren't any (to my knowledge) native
# Julia libraries to work with GIFs, but we have ImageMagick installed through BinDeps, which uses Homebrew
# since I'm on a Mac.  So let's just create a simple helper function to shell out to the `convert` binary.

# Inspired by Tom Breloff's animated plots: https://github.com/tbreloff/Plots.jl/blob/master/src/animation.jl
immutable GIF
    data::Vector{UInt8}
end
using Images
import Homebrew
"""
    animate(f, n; fps=20, width)

Call function `f` repeatedly, `n` times. The function `f` must take one argument (the frame number),
and it must return an Image for that frame.  Optionally specify the number of frames per second
and a width for proportional scaling (defaults to the actual width).
"""
function animate(f, n; fps = 20, width=0)
    mktempdir() do dir
        for i=1:n
            img = f(i)
            frame = width > 0 ? Images.imresize(img, (width, floor(Int, width/size(img, 1) * size(img, 2)))) : img
            Images.save(@sprintf("%s/%06d.png", dir, i), frame)
        end
        speed = round(Int, 100 / fps)
        run(`$(Homebrew.brew_prefix)/bin/convert -delay $speed -loop 0 $dir/*.png $dir/result.gif`)
        return GIF(open(readbytes, "$dir/result.gif"))
    end
end
Base.writemime(io::IO, ::MIME"text/html", g::GIF) = write(io, "<img src=\"data:image/gif;base64,$(base64encode(g.data))\" />")
Base.write(io::IO, g::GIF) = write(io, g.data)