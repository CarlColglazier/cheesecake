using LazyArrays
using Random


mutable struct Simulator23
	gd::GameData
	#automobile::Dict{Int,PredictionModel}
	autocharge::PredictionModel
	autoT::PredictionModel
	autoM::PredictionModel
	autoB::PredictionModel
	teleT::PredictionModel
	teleM::PredictionModel
	teleB::PredictionModel
	#endgame::Dict{Int,PredictionModel}
	dock::PredictionModel
end

function bymatch23(df::DataFrame)
	combine(
        filter(y -> length(y.key) == 6 && sum([y.score[1], y.score[6]]) != 1, groupby(df, [:key])),
		:key=>first=>:key,
        :team => (y -> [y[1:3]]) => :red,
        :team => (y -> [y[4:6]]) => :blue,
        :score => (y -> y[1]) => :red_score,
        :score => (y -> y[6]) => :blue_score,
        :winner => first => :winner,
        :match_number => first => :match_number,
        :time => first => :time,
		:activation => first => :red_activation,
		:activation => last => :blue_activation,
		:sustainability => first => :red_sustainability,
		:sustainability => last => :blue_sustainability
    )
end

"""
    piece_count_model(t::Matrix{Int}, y::Vector{Int}, priors::AbstractVector, N::Int)

Model for predicting the number of pieces on the field. 
	`t` is a 3xN matrix of teams;
	`y` is a vector of the observed number of pieces on the field;
	`priors`` is a vector of the prior means for each team; and
	`N` is the number of teams.
"""
@model function piece_count_model(t::Matrix{Int}, y::Vector{Int}, priors::Vector{Float64}, N::Int)
	ooff ~ Exponential(1)
	off ~ arraydist(truncated.(Normal.(priors,ooff); lower=0))
	lo = off[t[1,:]] + off[t[2,:]] + off[t[3,:]]
	y ~ arraydist(LazyArray(@~ Poisson.(lo)))
end

"""
    piece_count_model_tele(t::Matrix{Int}, y::Vector{Int}, auto_count::Vector{Int}, priors::AbstractVector, N::Int)

Model for predicting the number of pieces on the field, given a number of piece were already scored in autonomous mode.
	`t` is a 3xN matrix of teams;
	`y` is a vector of the observed number of pieces on the field;
	`auto_count` is a vector of the number of pieces scored in autonomous mode;
	`priors`` is a vector of the prior means for each team; and
	`N` is the number of teams.

Note: This currently does not account for the number of pieces scored in the automous period.
Trying to do so currently leaves the model in a state where it cannot converge.
"""
@model function piece_count_model_tele(t::Matrix{Int}, y::Vector{Int}, auto_count::Vector{Int}, priors::Vector{Float64}, N::Int)
	int ~ Normal(-3,1)
	ooff ~ Exponential(1)
	off ~ arraydist(Normal.(priors,ooff))
	lo = max.(off[t[1,:]], 0) + max.(off[t[2,:]], 0) + max.(off[t[3,:]], 0)
	y .~ BinomialLogit.(9, lo .+ int)
end

@model function dock_model2(docks::Vector{Int}, level::Vector{Bool}, t::Matrix, N::Int) # parked::Vector{Bool},
	eng_int ~ Normal(0, 0.25)
	balanced_int ~ Normal(2, 1)
	eng ~ filldist(Normal(0, 1), N)
	balanced ~ filldist(Normal(0, 1), N)
	#sum_e ~ Normal(0, 0.001 * N) # sum to zero
	#sum_e = sum(eng) + sum(balanced) + sum(park)
	for i in eachindex(docks)
		eng_sum = max.(eng[t[1,i]], 0) + max.(eng[t[2,i]], 0) + max.(eng[t[3,i]], 0)
		docks[i] ~ BinomialLogit(3, eng_sum + eng_int)
		if docks[i] > 0
			lev_av = (balanced[t[1,i]] + balanced[t[2,i]] + balanced[t[3,i]])
			level[i] ~ BernoulliLogit(lev_av + balanced_int)
		end
		#else
		#	park_sum = park[t[1,i]] + park[t[2,i]] + park[t[3,i]]
		#	parked[i] ~ BinomialLogit(3 - docks[i], park_sum + park_int)
		#end
	end
end

function build_dock_model(teams::Matrix, docked::Vector{Int}, parked::Vector{Int}, level::Vector{Bool}, teams_n::Int)
	logger = Logging.NullLogger()
	if length(docked) > 0
		model = dock_model2(docked, level, teams, teams_n)
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false);
		end
	else
		model = dock_model2([2],[true],[1 2 3;], 3)
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

"""
    simulate_links(n::Int)

Simulate the number of links that will be formed for a level, given `n` pieces are scored.
"""
function simulate_links(n::Int)
	# account for the fact these tend not to be random
	g = [Random.randperm(9) .<= n for _ in 1:5]
	return maximum(count_links.(g))
end


"""
    build_model23(gd::GameData, priors::Dict)

Build `Simulator23` for the given `GameData` and `priors`.
"""
function build_model23(gd::GameData, priors::Dict)
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

	basic_priors = [priors[y] for y in gd.teams]

	dock = build_dock_model(hcat(x.teams...), x.endgame_docked, x.endgame_parked, x.endgame_level, length(gd.teams))
    #team_auto_mobile = Dict{Int,PredictionModel}()
	#for team in keys(teams(gd))
	#	team_auto_mobile[team] = team_auto_mobile_model(gd, team)
	#end
	team_auto_charge = team_auto_charge_model23(gd, team_matrix, length(gd.teams))
	autoT = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countT, basic_priors, length(gd.teams)))
	autoM = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countM, basic_priors, length(gd.teams)))
	autoB = run_model(gd, piece_count_model(hcat(x.teams...), x.auto_countB, basic_priors, length(gd.teams)))
	
	teleT = run_model(gd, 
		piece_count_model_tele(
			hcat(x.teams...), 
			x.tele_countT,# .- x.auto_countT, 
			9 .- x.auto_countT,
			basic_priors, 
			length(gd.teams)
		)
	)
	teleM = run_model(gd, 
		piece_count_model_tele(
			hcat(x.teams...), 
			x.tele_countM,# .- x.auto_countM,
			9 .- x.auto_countM,
			basic_priors, 
			length(gd.teams)
		)
	)
	teleB = run_model(gd, 
		piece_count_model_tele(
			hcat(x.teams...), 
			x.tele_countB,# .- x.auto_countB,
			9 .- x.auto_countB,
			basic_priors, 
			length(gd.teams)
		)
	)

    return Simulator23(gd,
        #gd, team_auto_mobile, 
		team_auto_charge,
		autoT, autoM, autoB,
		teleT, teleM, teleB,
		dock
    )
end

@model function auto_charge(ye::Vector{Bool}, yd::Vector{Bool}, teams::Matrix, N::Int)
	e ~ filldist(Normal(0,1), N)
	eint ~ Normal(0,1)
	d ~ filldist(Normal(0,1), N)
	dint ~ Normal(0,1)
	eintsum = sum(eint)
	dintsum = sum(dint)
	le = max.(e[teams[:,1]], -1) + max.(e[teams[:,2]], -1) + max.(e[teams[:,3]], -1) .+ eintsum .+ eint
	ld = max.(d[teams[:,1]], -1) + max.(d[teams[:,2]], -1) + max.(d[teams[:,3]], -1) .+ dintsum .+ dint
	yd ~ arraydist(LazyArray(@~ BernoulliLogit.(ld)))
	ye ~ arraydist(LazyArray(@~ BernoulliLogit.(le)))
	#for i in eachindex(ye)
	#	#yd[i] ~ BernoulliLogit(ld[i] + dint)
	#	if yd[i]
	#		ye ~ BernoulliLogit(le[i])
	#	end
	#end
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
	eint = rand(first(get(chain, :eint)), n)
	dint = rand(first(get(chain, :dint)), n)
	docked = rand.(rand(BernoulliLogit.(dint), n))
	engaged = rand.(rand(BernoulliLogit.(eint), n))
	return max.(12 * engaged, 8 * docked)
end

function simulate_auto_charge_points(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.autocharge.chain
	team_indeces = [teams(sim.gd)[x] for x in teamsv]
	eint = first(get(chain, :eint))
	dint = first(get(chain, :dint))
	r_e = map(x -> max.(x, 0), first(get(chain, :e))[team_indeces])
	r_d = map(x -> max.(x, 0), first(get(chain, :d))[team_indeces])
	docked = rand.(rand(BernoulliLogit.(sum(r_d) .+ dint), n))
	engaged = rand.(rand(BernoulliLogit.(sum(r_e) .+ eint), n))
	return max.(12 * engaged, 8 * docked)
end

function simulate_auto_charge_points(sim::Simulator23, team::Int, n::Int)
	return simulate_auto_charge_points(sim, [team], n)
end

function endgame_level(r)
	egd = Dict("Docked"=>3, "Park"=>2, "None"=>1)
	if r.endGameBridgeState == "Level" && r.endgame_charge == "Docked"
		return 4
	else
		return egd[r.endgame_charge]
	end
end

function simulate_endgame(sim::Simulator23, team::Int, n::Int)
	chain = sim.dock.chain
	team_indices = teams(sim.gd)[team]
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	e = first(get(chain, :eng))[team_indices]
	b = first(get(chain, :balanced))[team_indices]
	docked = rand.(rand(BernoulliLogit.(eint .+ e), n))
	balanced = rand.(rand(BernoulliLogit.(bint .+ b), n))
	return docked, balanced
end

function simulate_endgame(sim::Simulator23, n::Int)
	chain = sim.dock.chain
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	docked = rand.(rand(BinomialLogit.(3, eint), n))
	balanced = rand.(rand(BernoulliLogit.(bint), n))
	return docked, balanced
end

function simulate_endgame(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.dock.chain
	eint = first(get(chain, :eng_int))
	bint = first(get(chain, :balanced_int))
	team_indeces = [teams(sim.gd)[x] for x in teamsv]
	eng_chains = first(get(chain, :eng))[team_indeces]
	e = sum(map(x -> max.(x, 0), eng_chains))
	b = sum(first(get(chain, :balanced))[team_indeces])
	docked = rand.(rand(BinomialLogit.(3, e .+ eint), n))
	balanced = rand.(rand(BernoulliLogit.(b .+ bint), n))
	return docked, balanced
end

function simulate_endgame_points(sim::Simulator23, n::Int)
	docked, balanced = simulate_endgame(sim, n)
	bal = balanced # TODO: is this how I want to do this?
	return 10 * docked .* bal .+ 6 * docked .* .!bal
end

function simulate_endgame_points(sim::Simulator23, team::Int, n::Int)
	docked, balanced = simulate_endgame(sim::Simulator23, team::Int, n::Int)
	return max.(10 * (docked .& balanced), 6 * docked)
end

# TODO: 
function simulate_endgame_points(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	docked, balanced = simulate_endgame(sim, teamsv, n)
	bal = balanced # TODO: is this how I want to do this?
	return 10 * docked .* bal .+ 6 * docked .* .!bal
end

function simulate_piece_counts(gd::GameData, pm::PredictionModel, n::Int; robots::Int=3)
	int = rand(median(Array(pm.chain[namesingroup(pm.chain, :off)]), dims=2), n)
	i = int .* (int .> 0)
	r = sum.(rand.(rand(first(get(pm.chain, :off)), n), robots))
	return min.(rand.(Poisson.(r .+ i)), 6)
end

function simulate_piece_counts(gd::GameData, pm::PredictionModel, teamsv::Vector{Int}, n::Int; robots::Int=3)
	team_indices = [teams(gd)[x] for x in teamsv]
	r_teams = rand(sum(first(get(pm.chain, :off))[team_indices]), n)
	other_r = sum.(rand.(rand(first(get(pm.chain, :off)), n), robots - length(teamsv)))
	ri = r_teams .+ other_r
	#ri = ri .* (ri .> 0)
	out = rand.(rand(Poisson.(ri), n))
	return min.(out, 6) # This can't ever go above 9, but we're assuming 6 is the most we're likely to see
end

function simulate_piece_counts_tele(gd::GameData, pm::PredictionModel, teamsv::Vector{Int}, auto_counts::Vector{Int}, n::Int)
	team_indices = [teams(gd)[x] for x in teamsv]
	int = first(get(pm.chain, :int))
	r_i = map(x -> max.(x, 0), first(get(pm.chain, :off))[team_indices])
	r = sum(r_i)
	ri = rand(int .+ r, n)
	return rand.(BinomialLogit.(9 .- auto_counts, ri))
end

function simulate_piece_counts_tele(pm::PredictionModel, auto_counts::Vector{Int}, n::Int)
	int = rand(first(get(pm.chain, :int)), n)
	return rand.(BinomialLogit.(9 .- auto_counts, int))
end

function simulate_teams_tuple(sim::Simulator23, teamsv::Vector{Int}, n)
	auto_charge=simulate_auto_charge_points(sim, teamsv, n)
	auto_countT=simulate_piece_counts(sim.gd, sim.autoT, teamsv, n)
	auto_countM=simulate_piece_counts(sim.gd, sim.autoM, teamsv, n)
	auto_countB=simulate_piece_counts(sim.gd, sim.autoB, teamsv, n)
	tele_countT=simulate_piece_counts_tele(sim.gd, sim.teleT, teamsv, auto_countT, n)
	tele_countM=simulate_piece_counts_tele(sim.gd, sim.teleM, teamsv, auto_countM, n)
	tele_countB=simulate_piece_counts_tele(sim.gd, sim.teleB, teamsv, auto_countB, n)
	link_count = simulate_links.(auto_countT .+ tele_countT) .+ simulate_links.(auto_countM .+ tele_countM) .+ simulate_links.(auto_countB .+ tele_countB)
	endgame=simulate_endgame_points(sim, teamsv, n)
	activation_scores=auto_charge .+ endgame
	activation = activation_scores .>=26
	sustainability=(link_count.>=rand([4,4,5],n))
	return (
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
	endgame=simulate_endgame_points(sim, n)
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


"""
    score(t::NamedTuple)

Converts a tuple of simulated results into a score.
"""
function score(t::NamedTuple)
	return (
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
	return Int.(round.(score(t) .- median(score(a))))
end

function simulate_teams(sim::Simulator23, teamsv::Vector{Int}, n)
	t = simulate_teams_tuple(sim, teamsv, n)
	return score(t)
end

"""
	simulate_teams(t::NamedTuple)

Returns the simulated scores for the simulation tuple `t`.
"""
function simulate_teams(t::NamedTuple)
	return score(t)
end

"""
	sim_evs(sim::Simulator23; n=100_000)

Returns a tuple of the teams and the EVs for each team.
"""
function sim_evs(sim::Simulator23; n=100_000)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, n) for y in x])
end

"""
    simulate_match(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)

Returns the simulated scores for the blue and red teams, respectively.
"""
function simulate_match(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim = simulate_teams_tuple(sim, blue, n)
	redsim = simulate_teams_tuple(sim, red, n)
	return bluesim, redsim
end

"""
    win_probabilities(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)

Returns the probability that the blue team wins, ties, and loses, respectively.
"""
function win_probabilities(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim, redsim = simulate_match(sim, blue, red; n=n)
	return [
		sum(simulate_teams(bluesim) .> simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .== simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .< simulate_teams(redsim)) / n,
	]
end