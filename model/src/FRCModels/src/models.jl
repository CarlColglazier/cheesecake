# Vector version is *much* faster
@model function count_model(t::Matrix{Int}, s::Vector{Int}, N::Int)
	μ_att ~ Normal(0.0, 0.1)
	ooff ~ Exponential(10)
	off ~ filldist(truncated(Normal(μ_att,ooff); lower=0), N)
	lo = log.(off[t[1,:]] + off[t[2,:]] + off[t[3,:]])
	return s ~ arraydist(LazyArray(@~ LogPoisson.(lo)))
end