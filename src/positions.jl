using StatsBase
using Images

immutable Position
    x::Float64
    y::Float64
    xspan::UnitRange{Int}
    yspan::UnitRange{Int}
    mass::Int
end

function positions(labels)
    N = maximum(labels)
    ps = Vector{Position}(N)
    for i=1:N
        mask = labels .== i
        xs = sum(mask, 2)
        ys = sum(mask, 1)
        ps[i] = Position(mean(1:length(xs), weights(xs)), mean(1:length(ys), weights(ys)),
                         findfirst(xs):findlast(xs), findfirst(ys):findlast(ys),
                         sum(xs))
    end
    ps
end

function filter_positions(labels, threshold=75)
    ps = positions(labels)
    filter(p->p.mass > threshold, ps)
end