using LazyArrays
using Random


mutable struct Simulator23
	gd::GameData
	autocharge::PredictionModel
	autoT::PredictionModel
	teleT::PredictionModel
	fouls::PredictionModel
	#dock::PredictionModel
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
@model function piece_count_model(t::Matrix{Int}, y::Vector{Int}, priors_mean::Vector{Float64}, priors_unc, N::Int)
	#ooff ~ Exponential(1)
	off ~ arraydist(truncated.(Normal.(priors_mean,priors_unc); lower=0))
	lo = off[t[1,:]] + off[t[2,:]] + off[t[3,:]]
	y ~ arraydist(LazyArray(@~ Poisson.(lo)))
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
	l = length(v)
	count = 0
	start = 1
	for i in eachindex(v)
		if v[i] == 0
			count += (i - start) ÷ 3
			start = i+1
		end
	end
	if length(v) >= l && v[l] == 1
		count += (l + 1 - start) ÷ 3
	end
	return count
end

"""
    simulate_links(n::Int)

Simulate the number of links that will be formed for a level, given `n` pieces are scored.
"""
function simulate_links(n::Int)
	# account for the fact these tend not to be random
	g = [Random.randperm(27) .<= n for _ in 1:10]
	return maximum(count_links.(g))
end


"""
    build_model23(gd::GameData, priors::Dict)

Build `Simulator23` for the given `GameData` and `priors`.
"""
function build_model23(gd::GameData, priors::Priors)
    x = combine(
        filter(x -> length(x.key) == 3, groupby(gd.df, [:event, :key, :alliance,])),
        :team => (x -> [[teams(gd)[y] for y in x]]) => :teams,
		:autoGamePieceCount => first => :autoGamePieceCount,
		:teleopGamePieceCount => first => :teleopGamePieceCount,
		:extraGamePieceCount => first => :extraGamePieceCount,
		:endgame_charge => (x -> sum(x .== "Docked")) => :endgame_docked,
		:endgame_charge => (x -> sum(x .== "Park")) => :endgame_parked,
		:endGameBridgeState => (x -> first(x) == "Level") => :endgame_level,
    )

	team_matrix = Matrix(reshape(gd.df.team, 3, size(gd.df,1) ÷ 3)')

	basic_priors = [priors.data["elos"][y][-1] for y in gd.teams]

	auto_priors_mean = copy(basic_priors)*2
	auto_priors_unc = fill(1.0, length(gd.teams))
	auto_priors_dict = get_priors(priors, "autoT", collect(gd.teams), gd.week)
	for (i, t) in enumerate(gd.teams)
		if haskey(auto_priors_dict, t)
			auto_priors_mean[i] = auto_priors_dict[t]
			auto_priors_unc[i] = 0.5
		end
	end

	tele_priors_mean = copy(basic_priors)*10
	tele_priors_unc = fill(2.0, length(gd.teams))
	tele_priors_dict = get_priors(priors, "teleT", collect(gd.teams), gd.week)
	for (i, t) in enumerate(gd.teams)
		if haskey(tele_priors_dict, t)
			tele_priors_mean[i] = tele_priors_dict[t]
			tele_priors_unc[i] = 1.0
		end
	end

	auto_charge_prior = fill(0.0, length(gd.teams))
	ac_priors_dict = get_priors(priors, "autocharge", collect(gd.teams), gd.week)
	for (i, t) in enumerate(gd.teams)
		if haskey(ac_priors_dict, t)
			auto_charge_prior[i] = (ac_priors_dict[t] |> x -> log(x/(1-x)))
		end
	end
	auto_count = x.autoGamePieceCount
	tele_count = x.teleopGamePieceCount + x.extraGamePieceCount

	#dock = build_dock_model(hcat(x.teams...), x.endgame_docked, x.endgame_parked, x.endgame_level, length(gd.teams))
	team_auto_charge = team_binary_model23(gd, team_matrix, auto_charge_prior, [y >= 8 for y in gd.df.auto_charge_points[begin:3:end]], length(gd.teams))
	autoT = run_model(gd, piece_count_model(hcat(x.teams...), auto_count, auto_priors_mean, auto_priors_unc, length(gd.teams)))
	teleT = run_model(gd, piece_count_model(hcat(x.teams...), max.(tele_count .- auto_count, 0), tele_priors_mean, tele_priors_unc, length(gd.teams)))
	team_fouls = team_binary_model23(gd, team_matrix, fill(0.0, length(gd.teams)), [y >= 1 for y in gd.df.foulCountFor[begin:3:end]], length(gd.teams))
    return Simulator23(gd,
		team_auto_charge,
		autoT,
		teleT,
		team_fouls
		#dock
    )
end

@model function binary_model(y::Vector{Bool}, teams::Matrix, priors::Vector, N::Int)
	d ~ arraydist(Normal.(priors, 1.0))
	dint ~ Normal(0,1)
	dintsum = sum(dint)
	ds = d[teams[:,1]] .+ d[teams[:,2]] .+ d[teams[:,3]] .+ dintsum .+ dint
	y ~ arraydist(LazyArray(@~ BernoulliLogit.(ds)))
end

function team_binary_model23(gd::GameData, t::Matrix, priors::Vector, y::Vector, team_n::Int)
	#df = gd.df
	logger = Logging.NullLogger()
	if nrow(gd.df) > 0
		model = binary_model(
			y,
			[teams(gd)[x] for x in t],
			priors,
			team_n
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 250; progress=false, verbose=false);
		end
	else
		model = binary_model([true, true],[true, true], [1 2 3; 4 5 6], fill(0.0, 6), 6)
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	probs = map.(x -> exp(x) / (exp(x) + 1), map(x -> x .+ first(get(s, :dint)), collect(first(get(s, :d)))))
	len_o = length(probs[1])
	auchargedf = DataFrame(
		:team=>collect(gd.teams),
		:mean=>mean.(probs),
		:low=>map(x -> _q(vec(x), Int(floor(0.05*len_o))), probs),
		:high=>map(x -> _q(vec(x), Int(ceil(0.95*len_o))), probs),
		:std=>std.(probs),
	)
	return PredictionModel(model, s, auchargedf)
end

function simulate_auto_charge_points(sim::Simulator23, n::Int)
	chain = sim.autocharge.chain
	dint = first(get(chain, :dint))
	docked = rand.(rand(BernoulliLogit.(dint), n))
	return 10 * docked
end

function simulate_auto_charge_points(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.autocharge.chain
	team_indeces = [teams(sim.gd)[x] for x in teamsv]
	dint = first(get(chain, :dint))
	pv = first(get(chain, :d))[team_indeces]

	m = sum(pv) .+ dint
	return 10 * rand.(rand(BernoulliLogit.(m), n))
end

function simulate_auto_charge_points(sim::Simulator23, team::Int, n::Int)
	return simulate_auto_charge_points(sim, [team], n)
end

function simulate_fouls(sim::Simulator23, n::Int)
	chain = sim.fouls.chain
	dint = first(get(chain, :dint))
	has_fouls = rand.(rand(BernoulliLogit.(dint), n))
	#return rand.(Exponential.(fill(0.5, n))) .* has_fouls
	return 1 * has_fouls
end

function simulate_fouls(sim::Simulator23, teamsv::Vector{Int}, n::Int)
	chain = sim.fouls.chain
	team_indeces = [teams(sim.gd)[x] for x in teamsv]
	dint = first(get(chain, :dint))
	pv = first(get(chain, :d))[team_indeces]
	m = sum(pv) .+ dint
	has_fouls = rand.(rand(BernoulliLogit.(m), n))
	#return rand.(Exponential.(fill(0.5, n))) .* has_fouls
	return 1 * has_fouls
end

function simulate_fouls(sim::Simulator23, team::Int, n::Int)
	return simulate_fouls(sim, [team], n)
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
	return rand.(Poisson.(r .+ i))
end

function simulate_piece_counts(gd::GameData, pm::PredictionModel, teamsv::Vector{Int}, n::Int; robots::Int=3)
	team_indices = [teams(gd)[x] for x in teamsv]
	pv = first(get(pm.chain, :off))[team_indices]
	m = map(Poisson, hcat(pv...))
	counts = sum(map(rand, m[rand(1:size(m,1), n), :]), dims=2)[:,1]
	#if length(teamsv) < robots
	#	counts += simulate_piece_counts(gd, pm, n; robots=robots-length(teamsv))
	#end
	return counts
end

function simulate_teams_tuple(sim::Simulator23, teamsv::Vector{Int}, n)
	auto_charge=simulate_auto_charge_points(sim, teamsv, n)
	auto_countT=simulate_piece_counts(sim.gd, sim.autoT, teamsv, n)
	tele_countT = simulate_piece_counts(sim.gd, sim.teleT, teamsv, n)
	link_count = simulate_links.(auto_countT .+ tele_countT)
	endgame=rand([0], n)#simulate_endgame_points(sim, teamsv, n)
	activation_scores=auto_charge .+ endgame
	activation = activation_scores .>=26
	sustainability=(link_count.>=rand([4,4,5],n))

	fouls = simulate_fouls(sim, teamsv, n)
	return (
		auto_charge=auto_charge,
		auto_countT=auto_countT,
		tele_countT=tele_countT,
		link_count=link_count,
		endgame=endgame,
		activation_scores=activation_scores,
		activation=activation,
		sustainability=sustainability,
		fouls=fouls,
	)
end

function simulate_teams_tuple(sim::Simulator23, n)
	auto_charge=simulate_auto_charge_points(sim, n)
	autoT=simulate_piece_counts(sim.gd, sim.autoT, n)
	teleT=simulate_piece_counts(sim.gd, sim.teleT, n)
	link_count = simulate_links.(autoT .+ teleT)# .+ simulate_links.(autoM .+ teleM) .+ simulate_links.(autoB .+ teleB)
	#endgame=simulate_endgame_points(sim, n)
	fouls = simulate_fouls(sim, n)
	return (
		#auto_mobile=rand([0,0,3], n), # TODO
		auto_charge=auto_charge,
		auto_countT=autoT,
		tele_countT=teleT,
		link_count=link_count,
		endgame=rand([0], n),
		fouls=fouls
		#endgame=endgame,
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
		5*t.auto_countT .+
		3.5*t.tele_countT .+
		0*t.endgame .+
		5*t.fouls
	)
end

function simulate_team(sim::Simulator23, team, n)
	t = simulate_team_tuple(sim, team, n)
	return score(t)
end

function ev_team(sim::Simulator23, team::Int, n::Int)
	t = simulate_team_tuple(sim, team, n)
	#a = simulate_teams_tuple(sim, n)
	#return Int.(round.(score(t) .- score(a)))
	return Int.(round.(score(t)))
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
	sim_evs(sim::Simulator23; n=1_000)

Returns a tuple of the teams and the EVs for each team.
"""
function sim_evs(sim::Simulator23; n=1_000)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, n) for y in x])
end

"""
    simulate_match(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 1_000)

Returns the simulated scores for the blue and red teams, respectively.
"""
function simulate_match(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 1_000)
	bluesim = simulate_teams_tuple(sim, blue, n)
	redsim = simulate_teams_tuple(sim, red, n)
	return bluesim, redsim
end

"""
    win_probabilities(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 1_000)

Returns the probability that the blue team wins, ties, and loses, respectively.
"""
function win_probabilities(sim::Simulator23, blue::Vector{Int}, red::Vector{Int}; n = 1_000)
	bluesim, redsim = simulate_match(sim, blue, red; n=n)
	return [
		sum(simulate_teams(bluesim) .> simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .== simulate_teams(redsim)) / n,
		sum(simulate_teams(bluesim) .< simulate_teams(redsim)) / n,
	]
end