using ArgParse
using Images
# using FixedPointNumbers, ImageMagick, Colors, Gadfly, DataFrames, ProgressMeter
# using Interpolations, AffineTransforms
import VideoIO

include("./rotate.jl")
include("./seek.jl")

function parse_commandline()
    s = ArgParseSettings()

    @add_arg_table s begin
        "--file"
            help = "filename"
        "--output"
            help = "output filename (x.png)"
            default = "../data/rotate-cropped.png"
        "--rotate", "-r"
            help = "rotation range"
            arg_type = Float64
            default = 0.0
        "--open"
            action = :store_true
            default = false
        "--x-min", "-x"
            arg_type = Int
            default = 0
        "--x-max", "-X"
            arg_type = Int
            default = 0
        "--y-min", "-y"
            arg_type = Int
            default = 0
        "--y-max", "-Y"
            arg_type = Int
            default = 0
        "--seek", "-s"
            help="seek in seconds"
            arg_type = Int
            default = 0
    end

    return parse_args(s)
end

function main()
    parsed_args = parse_commandline()
    
    io = VideoIO.open(parsed_args["file"])
    f = VideoIO.openvideo(io)
    
    seek(f, parsed_args["seek"])
    
    img = read(f, Image)

    # width:range, height:range 
    cropped = rotate_and_crop(img, parsed_args["rotate"], (parsed_args["x-min"]:parsed_args["x-max"], parsed_args["y-min"]:parsed_args["y-max"]))
    # isdir("../data") || mkdir("../data")
    Images.save(parsed_args["output"], cropped)
    if parsed_args["open"]
        run(`open $(parsed_args["output"])`)
    end
end

main()