using Images
import ImageMagick

function base64img(m::MIME, i::Image)
    try
        Images.save("/tmp/base64img.png", i)
        f = open("/tmp/base64img.png", "r")
        body = readall(f)
        close(f)
        return "data:$(m);base64,$(base64encode(body))"
    catch
        println("$(summary(i))")
        rethrow() 
    end
# this has issues w/ some Images so we won't use it
# 
#     buff = IOBuffer()
#     ImageMagick.writemime_(buff, MIME(m), i)
# #     writemime(buff, m, i, maxpixels=10^6)
#     body = base64encode(takebuf_string(buff))
#     return "data:$(m);base64,$(body)"
end
base64img(m::AbstractString, i::Image) = base64img(MIME(m), i)

# this should work, but breaks if you try to set maxpixel=10^7
# largewritemime(io::IO, m::AbstractString, x) = ImageMagick.writemime_(io, MIME(m), x)
# 
# function base64img(m::AbstractString, i)
#     return "data:$(m);base64,$(base64encode(largewritemime, m, i))"
# end