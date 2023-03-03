using LazyArrays
using Random


mutable struct Simulator23
	gd::GameData
	automobile::Dict{Int,PredictionModel}
	autocharge::PredictionModel
	autoT::PredictionModel
	autoM::PredictionModel
	autoB::PredictionModel
	teleT::PredictionModel
	teleM::PredictionModel
	teleB::PredictionModel
	endgame::Dict{Int,PredictionModel}
	dock::PredictionModel
end

@model function piece_count_model(t::Matrix{Int}, s::Vector{Int}, N::Int)
	int ~ Normal(0.0,1)
	ooff ~ Exponential(1)
	off ~ filldist(Normal(0,ooff), N)
	sum_e ~ Normal(0, 0.001 * N) # sum to zero
	sum_e = sum(off)
	lo = off[t[1,:]] + off[t[2,:]] + off[t[3,:]]
	for i in 1:length(s)
		s[i] ~ BinomialLogit(9, lo[i] .+ int)
	end
	return s
end

@model function piece_count_model_tele(t::Matrix{Int}, s::Vector{Int}, a::Vector{Int}, N::Int)
	int ~ Normal(0.0,1)
	ooff ~ Exponential(1)
	off ~ filldist(Normal(0,ooff), N)
	sum_e ~ Normal(0, 0.001 * N) # sum to zero
	sum_e = sum(off)
	lo = off[t[1,:]] + off[t[2,:]] + off[t[3,:]]
	s .~ BinomialLogit.(9 .- a, lo .+ int)
	return s
end

@model function endgame_model23(ye::Vector, yd::Vector, yp::Vector)
	e ~ Beta(1,2.0)
	d ~ Beta(1,1.5)
	p ~ Beta(1,1)
	ye .~ Bernoulli.(e)
	yd .~ Bernoulli.(d)
	yp .~ Bernoulli.(p)
	return e, d, p
end

@model function dock_model(y::AbstractVector, level::AbstractVector, parked::AbstractVector, t::Vector{Int}, N::Int)
	eng_int ~ Normal(0, 1)
	balanced_int ~ Normal(2, 1)
	park_int ~ Normal(2, 1)
	#ee ~ Exponential(5)
	sum_e ~ Normal(0, 0.001 * N) # sum to zero
	#
	eng ~ filldist(Normal(0, 1), N)
	balanced ~ filldist(Normal(0, 1), N)
	park ~ filldist(Normal(0, 1), N)

	sum_e = sum(eng)
	for i in eachindex(y)
		y[i] ~ BernoulliLogit(eng[t[i]] + eng_int)
		if y[i]
			level[i] ~ BernoulliLogit(balanced[t[i]] + balanced_int)
		else
			parked[i] ~ BernoulliLogit(park[t[i]] + park_int)
		end
	end
end

@model function dock_model2(docks::AbstractVector, level::AbstractVector, parked::AbstractVector, t::Matrix, N::Int)
	eng_int ~ Normal(0, 1)
	balanced_int ~ Normal(2, 1)
	park_int ~ Normal(2, 1)
	eng ~ filldist(Normal(0, 1), N)
	balanced ~ filldist(Normal(0, 1), N)
	park ~ filldist(Normal(0, 1), N)

	sum_e ~ Normal(0, 0.001 * N) # sum to zero
	sum_e = sum(eng) + sum(balanced) + sum(park)
	for i in eachindex(docks)
		eng_sum = eng[t[1,i]] + eng[t[2,i]] + eng[t[3,i]]
		docks[i] ~ BinomialLogit(3, eng_sum + eng_int)
		if docks[i] > 0
			lev_av = (balanced[t[1,i]] + balanced[t[2,i]] + balanced[t[3,i]])
			level[i] ~ BernoulliLogit(lev_av + balanced_int)
		else
			park_sum = park[t[1,i]] + park[t[2,i]] + park[t[3,i]]
			parked[i] ~ BinomialLogit(3 - docks[i], park_sum + park_int)
		end
	end
end

function build_dock_model(teams::Matrix, docked::Vector{Int}, parked::Vector{Int}, level::Vector{Bool}, teams_n::Int)
	logger = Logging.NullLogger()
	if length(docked) > 0
		model = dock_model2(docked, level, parked, teams, teams_n)
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false);
		end
	else
		model = dock_model2([2],[true],[1],[1 2 3;], 3)
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	endgamedf = DataFrame(
		#:team=>team,
	)
	return PredictionModel(model, s, endgamedf)
end

function count_links(v::AbstractVector)
	count = 0
	start = 1
	for i in eachindex(v)
		if v[i] == 0
			count += (i - start) ÷ 3
			start = i+1
		end
	end
	if length(v) >= 9 && v[9] == 1
		count += (10 - start) ÷ 3
	end
	return count
end

function simulate_links(n::Int)
	# account for the fact these tend not to be random
	g = [Random.randperm(9) .<= n for _ in 1:2]
	return maximum(count_links.(g))
end


function build_model23(gd::GameData)
    x = combine(
        filter(x -> length(x.key) == 3, groupby(gd.df, [:event, :key, :alliance,])),
        :team => (x -> [[teams(gd)[y] for y in x]]) => :teams,
		:auto_countT => first => :auto_countT,
		:auto_countM => first => :auto_countM,
		:auto_countB => first => :auto_countB,
		:teleop_countT => first => :tele_countT,
		:teleop_countM => first => :tele_countM,
		:teleop_countB => first => :tele_countB,
		:endgame_charge => (x -> sum(x .== "Docked")) => :endgame_docked,
		:endgame_charge => (x -> sum(x .== "Park")) => :endgame_parked,
		:endGameBridgeState => (x -> first(x) == "Level") => :endgame_level,
    )

	team_matrix = Matrix(reshape(gd.df.team, 3, size(gd.df,1) ÷ 3)')

    team_auto_mobile = Dict{Int,PredictionModel}()
	for team in keys(teams(gd))
		team_auto_mobile[team] = team_auto_mobile_model(gd, team)
	end

	team_auto_charge = team_auto_charge_model23(gd, team_matrix, length(gd.teams))

	endgame = Dict{Int,PredictionModel}()
    for team in keys(teams(gd))
        endgame[team] = team_endgame_model23(gd, team)
    end
	autoT = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countT, length(gd.teams)))
	autoM = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countM, length(gd.teams)))
	autoB = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countB, length(gd.teams)))
	teleT = run_model(gd, piece_count_model(hcat(x.teams...), x.tele_countT, length(gd.teams)))
	teleM = run_model(gd, piece_count_model(hcat(x.teams...), x.tele_countM, length(gd.teams)))
	teleB = run_model(gd, piece_count_model(hcat(x.teams...), x.tele_countB, length(gd.teams)))
	dock = build_dock_model(hcat(x.teams...), x.endgame_docked, x.endgame_parked, x.endgame_level, length(gd.teams))
    return Simulator23(
        gd, team_auto_mobile, team_auto_charge,
		autoT, autoM, autoB,
		teleT, teleM, teleB,
		endgame, dock
    )
end

@model function auto_mobility(y::Vector{Bool})
	b ~ Beta(1,1)
	y ~ Bernoulli(b)
end

# Generic?
function team_auto_mobile_model(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = auto_mobility(df.mobility)
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false)
		end
	else
		model = auto_mobility([false])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	modeldf = DataFrame(
		:team=>team,
		:auto_mobile=>mean(collect(first(get(s, :b))))
	)
	return PredictionModel(model, s, modeldf)
end

# Generic?
function simulate_auto_mobile(s::Chains, n::Int)
	return rand.(Bernoulli.(rand(first(get(s, :b)), n)))
end

@model function auto_charge(ye::Vector{Bool}, yd::Vector{Bool}, teams::Matrix, N::Int)
	e ~ filldist(Normal(0,1), N)
	eint ~ Normal(0,1)
	d ~ filldist(Normal(0,1), N)
	dint ~ Normal(0,1)
	le = e[teams[:,1]] + e[teams[:,2]] + e[teams[:,3]]
	ld = d[teams[:,1]] + d[teams[:,2]] + d[teams[:,3]]
	ye .~ BernoulliLogit.(le .+ eint)
	yd .~ BernoulliLogit.(ld .+ dint)
end

function team_auto_charge_model23(gd::GameData, t::Matrix, team_n::Int)
	df = gd.df
	logger = Logging.NullLogger()
	charge_points = gd.df.auto_charge_points[begin:3:end]
	if nrow(df) > 0
		model = auto_charge(
			[y >= 12 for y in charge_points],
			[y >= 8 for y in charge_points],
			[teams(gd)[x] for x in t], # need to have indices here.
			team_n
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false);
		end
	else
		model = auto_charge([true, true],[true, true], [1 2 3; 4 5 6], 6)
		#model = auto_charge([true, true],[true, true])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	
	auchargedf = DataFrame(
		:team=>collect(gd.teams),
		:engaged=>mean.(collect(first(get(s, :e)))),
		:docked=>mean.(collect(first(get(s, :d)))),
	)
	return PredictionModel(model, s, auchargedf)
end

function simulate_auto_charge_points(sim::Simulator23, n::Int)
	chain = sim.autocharge.chain
	eint = first(get(chain, :eint))
	dint = first(get(chain, :dint))
	r_eint = rand(eint, n)
	r_dint = rand(dint, n)
	return max.(12 * rand.(BernoulliLogit.(r_eint)), 8 * rand.(BernoulliLogit.(r_dint)))
end

function simulate_auto_charge_points(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.autocharge.chain
	eint = first(get(chain, :eint))
	dint = first(get(chain, :dint))
	r_eint = rand(eint, n)
	r_dint = rand(dint, n)
	r_e = []
	r_d = []
	for team in teamsv
		e = first(get(chain, :e))[teams(sim.gd)[team]]
		d = first(get(chain, :d))[teams(sim.gd)[team]]
		push!(r_e, rand(e, n))
		push!(r_d, rand(d, n))
	end
	return max.(12 * rand.(BernoulliLogit.(sum(r_e) .+ r_eint)), 8 * rand.(BernoulliLogit.(sum(r_d) .+ r_dint)))
end

function simulate_auto_charge_points(sim::Simulator23, team::Int, n::Int)
	return simulate_auto_charge_points(sim, [team], n)
end

egd = Dict("Engaged"=>4, "Docked"=>3, "Park"=>2, "None"=>1)
function endgame_level(r)
	if r.endGameBridgeState == "Level" && r.endgame_charge == "Docked"
		return 4
	else
		return egd[r.endgame_charge]
	end
end

function team_endgame_model23(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = endgame_model23(
			[endgame_level(r) >= 4 for r in eachrow(df)],
			[endgame_level(r) >= 3 for r in eachrow(df)],
			[endgame_level(r) >= 2 for r in eachrow(df)],
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false);
		end
	else
		model = endgame_model23([true],[true],[true])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	endgamedf = DataFrame(
		:team=>team,
		:engaged=>mean(collect(first(get(s, :e)))),
		:docked=>mean(collect(first(get(s, :d)))),
		:park=>mean(collect(first(get(s, :p)))),
	)
	return PredictionModel(model, s, endgamedf)
end

#=
function simulate_endgame_points(s::Chains, n::Int)
	levels = Dict(:e=>10, :d=>6, :p=>2)
	v = zeros(Int, 3, n)
	for (i, l) in enumerate([:e, :d, :p])
		v[i,:] = levels[l] * rand.(Bernoulli.(rand(first(get(s, l)), n)))
	end
	return maximum(v, dims=1)[1, :]
end
=#

function simulate_endgame(sim::Simulator23, team::Int, n::Int)
	chain = sim.dock.chain
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	pint = first(get(chain, :park_int))
	e = first(get(chain, :eng))[teams(sim.gd)[team]]
	b = first(get(chain, :balanced))[teams(sim.gd)[team]]
	p = first(get(chain, :park))[teams(sim.gd)[team]]
	docked = rand.(rand(BernoulliLogit.(eint .+ e), n))
	balanced = rand.(rand(BernoulliLogit.(bint .+ b), n))
	park = .!docked .& rand.(rand(BernoulliLogit.(pint .+ p), n))
	return docked, balanced, park
end

function simulate_endgame(sim::Simulator23, n::Int)
	chain = sim.dock.chain
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	pint = first(get(chain, :park_int))
	docked = rand.(rand(BinomialLogit.(3, eint), n))
	balanced = rand.(rand(BernoulliLogit.(bint), n))
	parked = rand.(BinomialLogit.(3 .- docked, rand(pint, n)))
	return docked, balanced, parked
end

function simulate_endgame(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.dock.chain
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	pint = first(get(chain, :park_int))
	team_indeces = [teams(sim.gd)[x] for x in teamsv]
	e = sum(first(get(chain, :eng))[team_indeces]) .- minimum(reduce(hcat, first(get(chain, :eng))[[1,2,3]]), dims=2)
	b = sum(first(get(chain, :balanced))[team_indeces])
	p = sum(first(get(chain, :park))[team_indeces])
	docked = rand.(rand(BinomialLogit.(3, e .+ eint), n))
	balanced = rand.(rand(BernoulliLogit.(b .+ bint), n))
	parked = rand.(BinomialLogit.(3 .- docked, rand(p .+ pint, n)))
	return docked, balanced, parked
end

function simulate_endgame_points(sim::Simulator23, n::Int)
	docked, balanced, park = simulate_endgame(sim, n)
	bal = balanced # TODO: is this how I want to do this?
	return 10 * docked .* bal .+ 6 * docked .* .!bal, 2 * sum(park)
end

function simulate_endgame_points(sim::Simulator23, team::Int, n::Int)
	docked, balanced, park = simulate_endgame(sim::Simulator23, team::Int, n::Int)
	return max.(10 * (docked .& balanced), 6 * docked), 2 * park
end

# TODO: 
function simulate_endgame_points(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	docked, balanced, park = simulate_endgame(sim, teamsv, n)
	bal = balanced # TODO: is this how I want to do this?
	return 10 * docked .* bal .+ 6 * docked .* .!bal, 2 * sum(park)
end

function simulate_piece_counts(gd::GameData, pm::PredictionModel, n::Int)
	int = rand(first(get(pm.chain, :int)), n)
	return rand.(BinomialLogit.(9, int))
end

function simulate_piece_counts(gd::GameData, pm::PredictionModel, teamsv::Vector{Int}, n::Int)
	int = rand(first(get(pm.chain, :int)), n)
	r = []
	for team in teamsv
		sp = first(get(pm.chain, :off))[teams(gd)[team]]
		push!(r, rand(sp, n))
	end
	return rand.(BinomialLogit.(Ref(9), sum(r) .+ int))
end

function simulate_piece_counts_tele(gd::GameData, pm::PredictionModel, teamsv::Vector{Int}, auto_counts::Vector{Int}, n::Int)
	int = rand(first(get(pm.chain, :int)), n)
	r = []
	for team in teamsv
		sp = first(get(pm.chain, :off))[teams(gd)[team]]
		push!(r, rand(sp, n))
	end
	return rand.(BinomialLogit.(9 .- auto_counts, int .+ sum(r)))
end

function simulate_piece_counts_tele(pm::PredictionModel, auto_counts::Vector{Int}, n::Int)
	int = rand(first(get(pm.chain, :int)), n)
	return rand.(BinomialLogit.(9 .- auto_counts, int))
end


function simulate_teams_tuple(sim::Simulator23, teamsv::Vector{Int}, n)
	#auto_mobile = 
	#endgame = simulate_endgame_points(sim, team, n)
	auto_charge=simulate_auto_charge_points(sim, teamsv, n)
	auto_countT=simulate_piece_counts(sim.gd, sim.autoT, teamsv, n)
	auto_countM=simulate_piece_counts(sim.gd, sim.autoM, teamsv, n)
	auto_countB=simulate_piece_counts(sim.gd, sim.autoB, teamsv, n)
	tele_countT=simulate_piece_counts_tele(sim.gd, sim.teleT, teamsv, auto_countT, n)
	tele_countM=simulate_piece_counts_tele(sim.gd, sim.teleM, teamsv, auto_countM, n)
	tele_countB=simulate_piece_counts_tele(sim.gd, sim.teleB, teamsv, auto_countB, n)
	link_count = simulate_links.(auto_countT .+ tele_countT) .+ simulate_links.(auto_countM .+ tele_countM) .+ simulate_links.(auto_countB .+ tele_countB)
	endgame, park=simulate_endgame_points(sim, teamsv, n)
	#eg_sim=deepcopy(endgame)
	activation_scores=auto_charge .+ endgame
	activation = activation_scores .>=26
	sustainability=link_count.>=rand([4,4,5],n)
	return (
		auto_mobile=sum([simulate_auto_mobile(sim.automobile[team].chain, n) for team in teamsv]),
		auto_charge=auto_charge,
		auto_countT=auto_countT,
		auto_countM=auto_countM,
		auto_countB=auto_countB,
		tele_countT=tele_countT,
		tele_countM=tele_countM,
		tele_countB=tele_countB,
		link_count=link_count,
		endgame=endgame,
		activation_scores=activation_scores,
		activation=activation,
		sustainability=sustainability
	)
end

function simulate_teams_tuple(sim::Simulator23, n)
	auto_charge=simulate_auto_charge_points(sim, n)
	autoT=simulate_piece_counts(sim.gd, sim.autoT, n)
	autoM=simulate_piece_counts(sim.gd, sim.autoM, n)
	autoB=simulate_piece_counts(sim.gd, sim.autoB, n) 
	teleT=simulate_piece_counts_tele(sim.teleT, autoT, n)
	teleM=simulate_piece_counts_tele(sim.teleM, autoM, n)
	teleB=simulate_piece_counts_tele(sim.teleB, autoT, n)
	link_count = simulate_links.(autoT .+ teleT) .+ simulate_links.(autoM .+ teleM) .+ simulate_links.(autoB .+ teleB)
	endgame, park=simulate_endgame_points(sim, n)
	return (
		auto_mobile=rand([0,0,3], n), # TODO
		auto_charge=auto_charge,
		auto_countT=autoT,
		auto_countM=autoM,
		auto_countB=autoB,
		tele_countT=teleT,
		tele_countM=teleM,
		tele_countB=teleB,
		link_count=link_count,
		endgame=endgame,
		#activation_scores=activation_scores,
		#activation=activation,
		#sustainability=sustainability
	)
end

function simulate_team_tuple(sim::Simulator23, team::Int, n)
	return simulate_teams_tuple(sim, [team], n)
end

function score(t::NamedTuple)
	return (
		3*t.auto_mobile .+
		t.auto_charge .+
		6*t.auto_countT .+
		4*t.auto_countM .+
		3*t.auto_countB .+
		5*t.tele_countT .+
		3*t.tele_countM .+
		2*t.tele_countB .+
		t.endgame
	)
end

function simulate_team(sim::Simulator23, team, n)
	t = simulate_team_tuple(sim, team, n)
	return score(t)
end

function ev_team(sim::Simulator23, team::Int, n::Int)
	t = simulate_team_tuple(sim, team, n)
	a = simulate_teams_tuple(sim, n)
	return Int.(round.(score(t) .- score(a)))
end

function simulate_teams(sim::Simulator23, teamsv::Vector{Int}, n)
	t = simulate_teams_tuple(sim, teamsv, n)
	return score(t)
end

function simulate_teams(t::NamedTuple)
	return score(t)
end

function sim_evs(sim::Simulator23; n=100_000)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, n) for y in x])
end

function run_event_once23(df, key)
	event_matches = df |>
		x -> x[x.event .== key, :] |> 
		x -> sort(x, :time)
	teams = OrderedSet(sort(event_matches.team))
	x = bymatch(event_matches)
	gddf = event_matches # completed?
	gd = GameData(gddf, teams)
	sim = build_model23(gd)
	return sim
end

function simulate_match(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim = simulate_teams_tuple(sim, blue, n) #.+ 4*simulate_count_teams(sim.gd, sim.fouls, red, n)
	redsim = simulate_teams_tuple(sim, red, n) #.+ 4*simulate_count_teams(sim.gd, sim.fouls, blue, n)
	return bluesim, redsim
end

function win_probabilities(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim, redsim = simulate_match(sim, blue, red; n=n)
	return [
		sum(simulate_teams(bluesim) .> simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .== simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .< simulate_teams(redsim)) / n,
	]
end