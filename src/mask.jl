using Images

# type Mask struct {
#   Start    int64 `json:"start,omitempty"`
#   End      int64 `json:"end,omitempty"`
#   BBox     BBox  `json:"bbox,omitempty"`
#   NullMask bool  `json:"null_mask,omitempty"`
# }

function mask(A::Image, masks)
    println("$(summary(A))")
    # a = convert(Image{ColorTypes.RGB{Float64}}, A)
    # a = convert(Image{ColorTypes.RGB{UFixed8}}, A)
    # T = eltype(A.data)
    # a = convert(Image{ColorTypes.RGB{T}}, A)
    # 1920x1080 Images.Image{ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8,8}},2,Array{ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8,8}},2}}
    fill = zero(eltype(A.data))
    # fill = zero(ColorTypes.RGB{UFixed8})
    for m in masks
        println("applying mask $m")
        if haskey(m, "null_mask")
            continue
        end
        if haskey(m, "start")
            A.data[:, m["start"]:m["end"]] = fill
        else
            bbox = m["bbox"]
            A.data[bbox["a"]["x"]:bbox["b"]["x"], bbox["a"]["y"]:bbox["b"]["y"]] = fill
        end
    end
    A
end 

# function mask{T}(A::BitMatrix, maskData)
# mask = imfilter_gaussian(grayim(abs((convert(Image{RGB{Float32}}, img) - convert(Image{RGB{Float32}}, background)).^2)),[3,3]) .> .06
# mask[:,13:14] = false
# mask[:,29:33] = false
# mask[:,40:43] = false
# mask[:,53:66] = false
# 
# grayim(map(UFixed8, mask))
 