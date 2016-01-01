using Images, Colors
import Base: .^
import FixedPointNumbers: UFixed8

# Absolute value is defined for RGB colors, but it's a little wonky -- it's the *sum* of the absolute values
# of the components. It is exactly what we want, but it's not defined for arrays of RGBs, so we add that definition here:
@vectorize_1arg AbstractRGB Base.abs

function (.^)(img::Image{RGB{Float32}}, pow::Integer)
    copy!(similar(img), reinterpret(RGB{Float32}, reinterpret(Float32, img).^pow))
end

function labelimg_base(img::Image, background::Image)
    i = grayim(abs( float32(img) - float32(background)))
    # https://github.com/JuliaIO/ImageMagick.jl/pull/18
    return map(Clamp01NaN(i), i)
    # grayim(convert(Image{ColorTypes.RGB{Float32}}, i))
    # grayim(map(ColorTypes.RGB{Float32}, i))
end

function label(img::Image, background::Image, blur=[3,3], tolerance=0.06)
    i = imfilter_gaussian(grayim(abs((convert(Image{RGB{Float32}}, img) - convert(Image{RGB{Float32}}, background)).^2)),blur) .> tolerance
    i::BitMatrix
    label_components(i) # This is like MATLAB's bwlabel
end

# """
# Show an image as it is being labeled
# """
# function labelimg(img::Image, background::Image, blur=[3,3], tolerance=0.06)
#     i = imfilter_gaussian(grayim(abs((convert(Image{RGB{Float32}}, img) - convert(Image{RGB{Float32}}, background)).^2)),blur) .> tolerance
#     return grayim(map(UFixed8, i))
# end

"""
Show an image with labeled (by color) regions
"""
function labelimg_example(img::Image, background::Image, blur=[3,3], tolerance=0.06)
    labels = label(img, background, blur, tolerance)
    # from https://github.com/JuliaGraphics/Colors.jl/blob/master/src/names_data.jl
    colors = [colorant"black", # this is the background
            colorant"red", colorant"yellow", colorant"green", colorant"blue", 
            colorant"orange", colorant"purple", colorant"gray", colorant"brown",
            colorant"aliceblue", colorant"burlywood", colorant"cadetblue", colorant"cyan",
            colorant"deeppink", colorant"gold", colorant"fuchsia", colorant"greenyellow",
            colorant"sandybrown", colorant"salmon", colorant"yellow4", colorant"violetred4"]
    Image(map(x->colors[x+1], labels'))
end