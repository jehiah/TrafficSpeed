using Images
import VideoIO
using ImageMagick
import FixedPointNumbers

import HttpServer: Server, Request, Response, HttpHandler
# https://github.com/JuliaWeb/HttpServer.jl
import JSON
# https://github.com/JuliaLang/JSON.jl
println("starting up")


include("./rotate.jl")
include("./seek.jl")
include("./base64img.jl")
include("./mask.jl")
include("./background.jl")

jsonContentType = Dict{AbstractString,AbstractString}([("Content-Type", "application/json")])

http = HttpHandler() do req::Request, res::Response
    println(req.resource)
    if ismatch(r"^/ping", req.resource)
        return Response(200, Dict{AbstractString,AbstractString}([("Content-Type", "text/plain")]), "OK")
    end
    
    if ismatch(r"^/debug/", req.resource)
        return Response(200, jsonContentType, JSON.json(req))
    end

    if ismatch(r"^/api/", req.resource)
        println("job is $(takebuf_string(IOBuffer(req.data)))")
        job = JSON.parse(IOBuffer(req.data))

        # println("job is $job")
        println("opening $(job["filename"])")

        resp = Dict{AbstractString,Any}()

        io = VideoIO.open(job["filename"])
        f = VideoIO.openvideo(io)
        resp["frames"] = length(f)
        resp["duration_seconds"] = duration(f)
        println("$(resp["frames"])frames, duration: $(resp["duration_seconds"]) seconds")

        img = read(f, Image)
        resp["video_resolution"] = "$(size(img.data, 1))x$(size(img.data, 2))"
        println("video resolution $(resp["video_resolution"])")
        # seek(f, job["seek"])

        println("Generating overview image (step 2)")
        resp["step_2_img"] = base64img("image/png", img)
        # println("img is $(resp["step_two_size"])")

        if haskey(job, "rotate") && job["rotate"] != 0.00001
            println("Rotating $(job["rotate"]) radians")
            # println("$(summary(img))")
            img = rotate(img, job["rotate"])
            println("$(summary(img))")
        end
        resp["step_3_img"] = base64img("image/png", img)

        if haskey(job, "bbox")
            println("cropping to $(job["bbox"])")
            # println("before crop $(summary(img))")
            job["bbox_region"] = (job["bbox"]["a"]["x"]:job["bbox"]["b"]["x"], job["bbox"]["a"]["y"]:job["bbox"]["b"]["y"])
            img = crop(img, job["bbox_region"])
            println("after crop $(summary(img))")
        else
            # set crop region to no-op size
            job["bbox_region"] = (1:size(img.data,1), 1:size(img.data, 2))
        end
        resp["cropped_resolution"] = "$(size(img.data, 1))x$(size(img.data, 2))"
        
        resp["step_4_img"] = base64img("image/png", img)
        if haskey(job, "masks")
            println("Applying masks: $(job["masks"])")
            masked = mask(img, job["masks"])
            resp["step_4_mask_img"] = base64img("image/png", masked)
        end
        
        # This gets called often, so let's optimize it a little bit.  Instead of just 
        # using read, I use the internal `retrieve!` with a pre-allocated buffer.
        # This is safe since I know it's getting rotated and discarded immediately
        seekstart(f)
        img = read(f, Image)
        const _rrc_buffer = Array{UInt8}(3, size(img.data, 1), size(img.data, 2))
        # inline read, rotate, crop w/ access to job
        function rrc(f::VideoIO.VideoReader)
            # _buffer is a 3-dimensional array (color x width x height), but by reinterpreting
            VideoIO.retrieve!(f, _rrc_buffer)
            if haskey(job, "rotate") && job["rotate"] != 0.00001
                Image(rotate_and_crop(reinterpret(ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8, 8}}, _rrc_buffer), job["rotate"], job["bbox_region"]), Dict("spatialorder"=>["x","y"]))
            else
                # just crop it (even if it's not really being cropped)
                Image(Base.unsafe_getindex(reinterpret(ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8, 8}}, _rrc_buffer), job["bbox_region"][1], job["bbox_region"][2]), Dict("spatialorder"=>["x","y"]))
            end
        end
        
        if job["step"] == "step_five"
            # generate a background
            println("Calculating background image")
            # background = rrc(f)
            background = avg_background(f, rrc)
        
            if haskey(job, "masks")
                background = mask(background, job["masks"])
            end

            resp["background_img"] = base64img("image/png", background)
        end

        Response(200, jsonContentType, JSON.json(resp))
    else
        Response(404, "404 - Not Found") 
    end
end

http.events["error"]  = (client, err) -> println(err)
http.events["listen"] = (port)        -> println("Listening on $port...")

function main()
    server = Server(http)
    run(server, 8000)
end

main()