@model function count_model_prior(t::Matrix{Int}, s::Vector{Int}, prior::Vector, var, N::Int)
	off ~ arraydist([truncated(Normal(prior[i], var); lower=0) for i in 1:N])
	lo = log.(off[t[1,:]] + off[t[2,:]] + off[t[3,:]])
	return s ~ arraydist(LazyArray(@~ LogPoisson.(lo)))
end

@model function endgame_model2(yt::Vector, yh::Vector, ym::Vector, yl::Vector)
	t ~ Beta(0.5,2)
	h ~ Beta(2,5)
	m ~ Beta(2,2)
	l ~ Beta(2,2)
	yt .~ Bernoulli.(t)
	yh .~ Bernoulli.(h)
	ym .~ Bernoulli.(m)
	yl .~ Bernoulli.(l)
	return t, h, m, l
end

@model function taxi(y::Vector{Bool})
	b ~ Beta(1.75,1)
	y ~ Bernoulli(b)
end