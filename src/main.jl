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
include("./labelimg.jl")
include("./positions.jl")

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
        video_summary(f)

        img = read(f, Image)
        resp["video_resolution"] = "$(size(img.data, 1))x$(size(img.data, 2))"
        println("video resolution $(resp["video_resolution"])")
        # seek(f, job["seek"])

        println("Generating overview image (step 2)")
        resp["step_2_img"] = base64img("image/png", img)

        if haskey(job, "rotate") && job["rotate"] != 0.00001
            println("Rotating $(job["rotate"]) radians")
            # println("$(summary(img))")
            img = rotate(img, job["rotate"])
            println("$(summary(img))")
        end
        resp["step_3_img"] = base64img("image/png", img)

        if haskey(job, "bbox")
            println("cropping to [$(job["bbox"]["a"]["x"]):$(job["bbox"]["b"]["x"]), $(job["bbox"]["a"]["y"]):$(job["bbox"]["b"]["y"])]")
            # println("before crop $(summary(img))")
            job["bbox_region"] = (job["bbox"]["a"]["x"]:job["bbox"]["b"]["x"], job["bbox"]["a"]["y"]:job["bbox"]["b"]["y"])
            img = crop(img, job["bbox_region"])
            println("Cropped image is $(summary(img))")
        else
            # set crop region to no-op size
            job["bbox_region"] = (1:size(img.data,1), 1:size(img.data, 2))
        end
        resp["cropped_resolution"] = "$(size(img.data, 1))x$(size(img.data, 2))"
        
        resp["step_4_img"] = base64img("image/png", img)
        if haskey(job, "masks")
            # println("Applying masks: $(job["masks"])")
            masked = mask_img(img, job["masks"])
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
        
            resp["background_img"] = base64img("image/png", background)
            
            # pick five frames
            frame_analysis = Array{Any, 1}()
            i = 0
            blur_arg=[job["blur"], job["blur"]]
            if haskey(job, "masks")
                mask_args = job["masks"]
            else
                mask_args = Array{Any,1}()
            end

            while i < 4
                e = Dict{AbstractString,Any}()
                pos = floor(Int, i * (duration(f)/5)) # increment by a smaller fraction so we don't get the last frame
                println("analyzing frame at $pos seconds")
                e["ts"] = pos
                seek(f, pos)
                frame = rrc(f)

                e["highlight"] = base64img("image/png", labelimg_base(frame, background))
                e["colored"] = base64img("image/png", labelimg_example(frame, background, mask_args, blur_arg, job["tolerance"]))
                e["positions"] = positions(label(frame, background, mask_args, blur_arg, job["tolerance"]))
                # println("$i positions json is $(JSON.json(e["positions"]))")
                
                push!(frame_analysis, e)
                i += 1
            end
            resp["frame_analysis"] = frame_analysis
            
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