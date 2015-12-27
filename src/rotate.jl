using Interpolations, AffineTransforms, Images
import AffineTransforms: center

import VideoIO

"""
Rotate and crop a matrix by the angle θ.

Optional arguments:
* region - a tuple of two arrays that specify the section of the rotated image to return; defaults to the unrotated viewport
* fill - the value to use for regions that fall outside the rotated image; defaults to zero(T)
"""

function rotate{T}(A::AbstractMatrix{T}, θ, fill=zero(T))
    # taken from https://github.com/timholy/AffineTransforms.jl/blob/master/test/runtests.jl#L178
    itp = interpolate(A, BSpline(Linear()), OnGrid())
    tfm = tformrotate(θ)
    tA = AffineTransforms.TransformedArray(extrapolate(itp, fill), tfm)
    dest = AffineTransforms.transform(tA)
    tfm_recentered = AffineTransforms.AffineTransform(tfm.scalefwd, tfm.offset + center(A) - tfm.scalefwd*center(dest))
    tA_recentered = AffineTransforms.TransformedArray(extrapolate(itp, fill), tfm_recentered)
    tA_recentered
end
rotate(A::Image, θ) = copyproperties(A, rotate(A.data, θ))

function rotate_and_crop{T}(A::AbstractMatrix{T}, θ, region=(1:size(A, 1), 1:size(A, 2)), fill=zero(T))
    etp = extrapolate(interpolate(A, BSpline(Linear()), OnGrid()), fill)
    R = TransformedArray(etp, tformrotate(θ))
    if region[1][1] == 0 && region[2][1] == 0
        region = (1:size(R, 1), 1:size(R, 2))
    end
    Base.unsafe_getindex(R, region[1], region[2]) # Extrapolations can ignore bounds checks
end

# While the above will work for images, it may iterate through them inefficiently depending on the storage order
rotate_and_crop(A::Image, θ, region) = shareproperties(A, rotate_and_crop(A.data, θ, region))

# # This gets called often, so let's optimize it a little bit.  Instead of just 
# # using read, I use the internal `retrieve!` with a pre-allocated buffer.
# # This is safe since I know it's getting rotated and discarded immediately
# const _buffer = Array{UInt8}(3, size(img.data, 1), size(img.data, 2))
# function readroi(f::VideoIO.VideoReader, region=(1:size(A, 1), 1:size(A, 2)))
#     VideoIO.retrieve!(f, _buffer)
#     # _buffer is a 3-dimensional array (color x width x height), but by reinterpreting
#     # it as RGB{UFixed8}, it becomes a matrix of colors that we can rotate
#     Image(rotate_and_crop(reinterpret(RGB{UFixed8}, _buffer), 0.321, region), Dict("spatialorder"=>["x","y"]))
# end