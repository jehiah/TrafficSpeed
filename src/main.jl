using Images
import VideoIO
using ImageMagick

import HttpServer: Server, Request, Response, HttpHandler
# https://github.com/JuliaWeb/HttpServer.jl
import JSON
# https://github.com/JuliaLang/JSON.jl
println("starting up")


include("./rotate.jl")
include("./seek.jl")
include("./base64img.jl")

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
        println("jsondata is $(takebuf_string(IOBuffer(req.data)))")
        job = JSON.parse(IOBuffer(req.data))

        println("job is $job")
        println("opening $(job["filename"])")

        resp = Dict{AbstractString,Any}()

        io = VideoIO.open(job["filename"])
        f = VideoIO.openvideo(io)
        resp["frames"] = length(f)
        resp["duration_seconds"] = duration(f)

        # seek(f, job["seek"])
        println("step_2_img")
        img = read(f, Image)
        resp["step_2_img"] = base64img("image/png", img)
        resp["step_2_size"] = "$(size(img.data, 1))x$(size(img.data, 2))"
        # println("img is $(resp["step_two_size"])")

        if haskey(job, "rotate")
            println("rotating $(job["rotate"])")
            img = rotate(img, job["rotate"])
        end
        println("step_3_img")
        resp["step_3_img"] = base64img("image/png", img)
        if haskey(job, "bbox")
            println("cropping $(job["bbox"])")
            # println("before $(size(cropped.data, 1))x$(size(cropped.data, 2))")
            img = sliceim(img, "x", job["bbox"]["a"]["x"]:job["bbox"]["b"]["x"], "y", job["bbox"]["a"]["y"]:job["bbox"]["b"]["y"])
            # cropped = crop(cropped, (parsed_args["x-min"]:parsed_args["x-max"], parsed_args["y-min"]:parsed_args["y-max"]))
            # println("after $(size(cropped.data, 1))x$(size(cropped.data, 2))")
        end
        println("cropping step_4_img")
        resp["step_4_img"] = base64img("image/png", img)

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