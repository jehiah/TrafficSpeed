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
include("./gif.jl")

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
        resp["overview_img"] = base64img("image/png", img)
        gif = animate(15, fps=15, width=400) do _
            read(f, Image) # throw 50% away
            read(f, Image)
        end
        resp["overview_gif"] = base64gif(gif)

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
            VideoIO.read!(f, _rrc_buffer)
            # read!(f, img)
            # buffer = read(f, Image)
            if haskey(job, "rotate") && job["rotate"] != 0.00001
                # Image(rotate_and_crop(buffer, job["rotate"], job["bbox_region"]), Dict("spatialorder"=>["x","y"]))
                Image(rotate_and_crop(reinterpret(ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8, 8}}, _rrc_buffer), job["rotate"], job["bbox_region"]), Dict("spatialorder"=>["x","y"]))
            else
                # just crop it (even if it's not really being cropped)
                # Image(Base.unsafe_getindex(buffer, job["bbox_region"][1], job["bbox_region"][2]), Dict("spatialorder"=>["x","y"]))
                Image(Base.unsafe_getindex(reinterpret(ColorTypes.RGB{FixedPointNumbers.UFixed{UInt8, 8}}, _rrc_buffer), job["bbox_region"][1], job["bbox_region"][2]), Dict("spatialorder"=>["x","y"]))
            end
        end
        
        if job["step"] >= 5
            # generate a background
            println("Calculating background image")
            if job["step"] == 5 
                background = avg_background(f, rrc, 25)
            else
                # faster for now
                background = avg_background(f, rrc, 10)
            end
  
            resp["background_img"] = base64img("image/png", background)
        end
        if job["step"] == 5
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
                e["base"] = base64img("image/png", frame)
                # seek(f, pos)
                # gif = animate(15, fps=15, width=400) do _
                #     read(f, Image) # throw away 50% of frames
                #     rrc(f)
                # end
                # e["base_gif"] = base64gif(gif)

                seek(f, pos)
                e["highlight"] = base64img("image/png", labelimg_base(frame, background))
                gif = animate(15, fps=15, width=400) do _
                    read(f, Image) # throw away 50% of frames
                    labelimg_base(rrc(f), background)
                end
                e["highlight_gif"] = base64gif(gif)


                e["colored"] = base64img("image/png", labelimg_example(frame, background, mask_args, blur_arg, job["tolerance"]))
                e["positions"] = positions(label(frame, background, mask_args, blur_arg, job["tolerance"]))
                # println("$i positions json is $(JSON.json(e["positions"]))")

                seek(f, pos)
                gif = animate(15, fps=15, width=400) do _
                    read(f, Image) # throw away 50% of frames
                    labelimg_example(rrc(f), background, mask_args, blur_arg, job["tolerance"])
                end
                e["colored_gif"] = base64gif(gif)

                push!(frame_analysis, e)
                i += 1
            end
            resp["frame_analysis"] = frame_analysis
        end

        if job["step"] == 6
            # we just need to eco back a frame
            if haskey(job, "seek")
                seek(f, job["seek"])
                resp["step_6_img"] = base64img("image/png", rrc(f))
            end
            # extract an image for each calibration
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