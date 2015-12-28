using HttpServer
# https://github.com/JuliaWeb/HttpServer.jl
import JSON
# https://github.com/JuliaLang/JSON.jl

http = HttpHandler() do req::Request, res::Response
    println(req.resource)
    if ismatch(r"^/api/", req.resource)
        # JSON.parse(req.data)
        Response(200, Dict{AbstractString,AbstractString}([("Content-Type", "text/plain")]), JSON.json(req))
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