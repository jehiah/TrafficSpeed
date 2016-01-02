using Images

# type Mask struct {
#   Start    int64 `json:"start,omitempty"`
#   End      int64 `json:"end,omitempty"`
#   BBox     BBox  `json:"bbox,omitempty"`
#   NullMask bool  `json:"null_mask,omitempty"`
# }

function mask_img(A::Image, masks::Array)
    fill = zero(eltype(A.data))
    for m in masks
        if haskey(m, "null_mask")
            continue
        end
        if haskey(m, "start")
            println("masking [:, $(m["start"]):$(m["end"])]")
            A.data[:, m["start"]:m["end"]] = fill
        else
            bbox = m["bbox"]
            println("masking [$(bbox["a"]["x"]):$(bbox["b"]["x"]), $(bbox["a"]["y"]):$(bbox["b"]["y"])]")
            A.data[bbox["a"]["x"]:bbox["b"]["x"], bbox["a"]["y"]:bbox["b"]["y"]] = fill
        end
    end
    A
end 

function mask(A::BitMatrix, masks::Array)
    for m in masks
        if haskey(m, "null_mask")
            continue
        end
        if haskey(m, "start")
            A[:, m["start"]:m["end"]] = false
        else
            bbox = m["bbox"]
            A[bbox["a"]["x"]:bbox["b"]["x"], bbox["a"]["y"]:bbox["b"]["y"]] = false
        end
    end
    A
end
