module FRCModels

using DataFrames
using Arrow
using CSV
using ReverseDiff
using LinearAlgebra
using Logging
using LazyArrays
using Turing
import OrderedCollections: OrderedSet

#using ReverseDiff

#Turing.setadbackend(:reversediff)
#Turing.setrdcache(true)

struct GameData
	df::DataFrame
	teams::OrderedSet{Int}
end

struct PredictionModel
	model::Turing.DynamicPPL.Model
	chain::Turing.Chains
	summary::DataFrame
end

struct Simulator
	gd::GameData
	taxi::Dict{Int,PredictionModel}
	cargoupper::PredictionModel
	cargolower::PredictionModel
	cargoautoupper::PredictionModel
	cargoautolower::PredictionModel
	climb::Dict{Int, PredictionModel}
	fouls::PredictionModel
end

Simulator(d::Dict{String, Any}) = Simulator(
    d["gd"], d["cargoupper"], d["cargolower"], d["cargoautoupper"], d["cargoautolower"], d["climb"], d["fouls"]
)

function bymatch(df::DataFrame)
	combine(
        filter(y -> length(y.key) == 6 && sum([y.score[1], y.score[6]]) != 1, groupby(df, [:key])),
        :team => (y -> [y[1:3]]) => :red,
        :team => (y -> [y[4:6]]) => :blue,
        :score => (y -> y[1]) => :red_score,
        :score => (y -> y[6]) => :blue_score,
        :winner => first => :winner,
        :match_number => first => :match_number,
        :time => first => :time,
    )
end

"""
    samples_df(s, teamsdict)

    Create a summary DataFrame from chain data.
"""
function samples_df(s::Turing.Chains, teamsdict::Dict{Int, Int})
	off = collect(first(get(s, :off)))
	return DataFrame(
		:team=>[teamsdict[x] for x in 1:length(off)],
		:mean=>mean.(off),
		:std=>std.(off),
	)
end

function teams(gd::GameData)
	return  gd.teams |> x -> Dict(x.=>1:length(x))
end

function teamsr(gd::GameData)
	return  gd.teams |> x -> Dict(1:length(x).=>x)
end

"""
    _q(v::Vector, n::Int)

    Get the `n`th element in `v`.
"""
function _q(v::Vector, n::Int)
	return sort(v)[n]
end


# Vector version is *much* faster
@model function count_model(t::Matrix{Int}, s::Vector{Int}, N::Int)
	μ_att ~ Normal(0.0, 0.1)
	ooff ~ Exponential(10)
	off ~ filldist(truncated(Normal(μ_att,ooff); lower=0), N)
	lo = log.(off[t[1,:]] + off[t[2,:]] + off[t[3,:]])
	return s ~ arraydist(LazyArray(@~ LogPoisson.(lo)))
end

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

function run_model(gd::GameData, model; fast=false)
	if fast
		sampler = HMC(0.05, 1)
		samples = 1000
	else
		sampler = NUTS()
		samples = 100
	end
	logger = Logging.NullLogger()
	s = Logging.with_logger(logger) do
		sample(model, sampler, samples; drop_warmup=true, progress=false, verbose=false)
	end
	summary = samples_df(s, teamsr(gd))
	sort!(summary, :mean; rev=true)
	return PredictionModel(model, s, summary)
end

function build_model(gd::GameData, elos)
    x = combine(
        filter(x -> length(x.key) == 3, groupby(gd.df, [:event, :key, :alliance,])),
        :team => (x -> [[teams(gd)[y] for y in x]]) => :teams,
        :teleopCargoUpper => first => :cargoTeleUpper,
        :autoCargoUpper => first => :cargoAutoUpper,
        :teleopCargoLower => first => :cargoTeleLower,
        :autoCargoLower => first => :cargoAutoLower,
        :foulCount => first => :foulCount,
        :techFoulcount => first => :techFoulCount,
    )

	var_prior = 1

    # These take a while
    cargoupper = Threads.@spawn run_model(
        gd,
        count_model_prior(
            hcat(x.teams...),
            x.cargoTeleUpper,
            [minimum([exp((elos[teamsr(gd)[x]] - 1550) / 100), 10]) for x in 1:length(gd.teams)],
            var_prior * 10, # var prior
            length(gd.teams)
        ); fast=false
    )

    cargolower = Threads.@spawn run_model(
        gd,
        count_model(
            hcat(x.teams...),
            x.cargoTeleLower,
            length(gd.teams)
        )
    )

    cargoautoupper = Threads.@spawn run_model(
        gd,
        count_model_prior(
            hcat(x.teams...),
            x.cargoAutoUpper,
            [minimum([0.5 * exp((elos[x] - 1600) / 100), 2.5]) for x in gd.teams],
            var_prior * 2.5, # var prior
            length(teams(gd))
        ); fast=false
    )

    cargoautolower = Threads.@spawn run_model(
        gd,
        count_model(
            hcat(x.teams...),
            x.cargoAutoLower,
            length(teams(gd))
        ); fast=false
    )

    fouls = Threads.@spawn run_model(
        gd,
        count_model(
            hcat(x.teams...),
            x.foulCount,
            length(teams(gd))
        ); fast=false
    )

    # Endgame
    team_climbs = Dict{Int,PredictionModel}()
    for team in keys(teams(gd))
        team_climbs[team] = team_climb_model(gd, team)
    end
    #
	team_taxi = Dict{Int,PredictionModel}()
	for team in keys(teams(gd))
		team_taxi[team] = team_taxi_model(gd, team)
	end
    return Simulator(
        gd, team_taxi, fetch(cargoupper), fetch(cargolower),
        fetch(cargoautoupper), fetch(cargoautolower),
        team_climbs, fetch(fouls)
    )
end

function team_taxi_model(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = taxi(df.taxi)
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 100; progress=false, verbose=false)
		end
	else
		model = taxi([false])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	modeldf = DataFrame(
		:team=>team,
		:taxi=>mean(collect(first(get(s, :b))))
	)
	return PredictionModel(model, s, modeldf)
end


egd = Dict("Traversal"=>5, "High"=>4, "Mid"=>3, "Low"=>2, "None"=>1)
function team_climb_model(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = endgame_model2(
			[egd[y] >= 5 for y in df.endgame],
			[egd[y] >= 4 for y in df.endgame],
			[egd[y] >= 3 for y in df.endgame],
			[egd[y] >= 2 for y in df.endgame]
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 100; progress=false, verbose=false);
		end
	else
		model = endgame_model2([true],[true],[true],[true])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	endgamedf = DataFrame(
		:team=>team,
		:traversal=>mean(collect(first(get(s, :t)))),
		:high=>mean(collect(first(get(s, :h)))),
		:mid=>mean(collect(first(get(s, :m)))),
		:low=>mean(collect(first(get(s, :l)))),
	)
	return PredictionModel(model, s, endgamedf)
end

function simulate_counts(gd::GameData, pm::PredictionModel, team, n)
	sp = first(get(pm.chain, :off))[teams(gd)[team]]
	#rd = truncated(Normal(mean(sp), std(sp)); lower=0)
	#return rand.(Poisson.(rand(rd, n)))
	r = rand(sp, n)
	return rand.(Poisson.(r))
end

function simulate_taxi(s::Chains, n::Int)
	return rand.(Bernoulli.(rand(first(get(s, :b)), n)))
end

function simulate_climb_points(s::Chains, n::Int)
	levels = Dict(:t=>15, :h=>10, :m=>6, :l=>4)
	v = zeros(Int, 4, n)
	for (i, l) in enumerate([:t, :h, :m, :l])
		v[i,:] = levels[l] * rand.(Bernoulli.(rand(first(get(s, l)), n)))
	end
	return maximum(v, dims=1)[1, :]
end

function simulate_climb_points(sim::Simulator, team::Int, n::Int)
	chain = sim.climb[team].chain
	return simulate_climb_points(chain, n)
end

function simulate_count_teams(gd::GameData, pm::PredictionModel, teams::Vector{Int}, n::Int)
	return sum(simulate_counts.(Ref(gd), Ref(pm), teams, n))
end

function simulate_total_cargo(sim::Simulator, teams::Vector{Int}, n::Int)
	return (
		simulate_count_teams(sim.gd, sim.cargoautolower, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargoautoupper, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargoupper, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargolower, teams, n)
	)
end

#=
function simulate_team(sim::Simulator, team, n)
	return (4*simulate_counts(sim.gd, sim.cargoautoupper, team, n) .+
			2*simulate_counts(sim.gd, sim.cargoautolower, team, n) .+
			2*simulate_counts(sim.gd, sim.cargoupper, team, n) .+
			simulate_counts(sim.gd, sim.cargolower, team, n) .+
			simulate_climb_points(sim, team, n)
	)
end
=#

function simulate_team_tuple(sim::Simulator, team, n)
	cargoautoupper = simulate_counts(sim.gd, sim.cargoautoupper, team, n)
	cargoautolower = simulate_counts(sim.gd, sim.cargoautolower, team, n)
	cargoupper = simulate_counts(sim.gd, sim.cargoupper, team, n)
	cargolower = simulate_counts(sim.gd, sim.cargolower, team, n)
	climb = simulate_climb_points(sim, team, n)
	taxi = simulate_taxi(sim.taxi[team].chain, n)
	return (
		cargoautoupper=cargoautoupper,
		cargoautolower=cargoautolower,
		cargoupper=cargoupper,
		cargolower=cargolower,
		climb=climb,
		taxi=taxi
	)
end

function simulate_team(sim::Simulator, team, n)
	t = simulate_team_tuple(sim, team, n)
	return (
		4*t.cargoautoupper .+
		2*t.cargoautolower .+
		2*t.cargoupper .+
		t.cargolower .+
		t.climb .+
		4*t.taxi
	)
end

function ev_team(sim::Simulator, team::Int, n::Int)
	# TODO: Fouls
	return simulate_team(sim, team, n) #.- 4*simulate_counts(sim.gd, sim.fouls, team, n)
end

function simulate_teams(sim::Simulator, teams::Vector{Int}, n)
	return sum(
		[simulate_team(sim, x, n) for x in teams]
	)
end

function win_probabilities(sim::Simulator, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim = simulate_teams(sim, blue, n) .+ 4*simulate_count_teams(sim.gd, sim.fouls, red, n)
	redsim = simulate_teams(sim, red, n) .+ 4*simulate_count_teams(sim.gd, sim.fouls, blue, n)
	return [
		sum(bluesim .> redsim) / n,
		sum(bluesim .== redsim) / n,
		sum(bluesim .< redsim) / n,
	]
end

function run_event_once(df, key, elos)
	event_matches = df |>
		x -> x[x.event .== key, :] |> 
		x -> sort(x, :time)
	teams = OrderedSet(sort(event_matches.team))
	x = bymatch(event_matches)
	gddf = event_matches # completed?
	gd = GameData(gddf, teams)
	sim = build_model(gd, elos)
	return sim
end

#### Run the model for an entire event!
function run_event(df, key, elos)
	event_matches = df |>
		x -> x[x.event .== key, :] |> 
		x -> sort(x, :time)

	teams = OrderedSet(sort(event_matches.team))
	predictions = []
	#simulations = []
	x = bymatch(event_matches)
	for (i, r) in enumerate(eachrow(x))
		println(i)
		if i <= 1
			push!(predictions, 0.5)
			continue
		end
		gddf = event_matches |> x -> x[x.time .< r.time, :]
		gd = GameData(gddf, teams)
		sim = build_model(gd, elos);
		p = win_probabilities(sim, r.red, r.blue)
		push!(predictions, first(p))
		#push!(predictions, [r.key, first(p), r.winner .== "red"])
		#push!(simulations, sim)
	end
	return predictions#, simulations
end

function sim_evs(sim::Simulator)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, 100_000) for y in x])
end

function evs(sim::Simulator)
	return sim_evs(sim) |>
	    x -> DataFrame(
		    :team => x[:teams],
		    :ev_median => median.(x[:sims]),
		    :ev_low => _q.(x[:sims], Int(100_000*0.05)),
            :ev_high => _q.(x[:sims], Int(100_000*0.95))
		) |>
		x -> sort(x, :ev_median; rev=true)
end

end # module FRCModels
