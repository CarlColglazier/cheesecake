mutable struct Simulator23
	gd::GameData
	automobile::Dict{Int,PredictionModel}
	autocharge::Dict{Int,PredictionModel}
	autoConeT::PredictionModel
	autoConeM::PredictionModel
	autoConeB::PredictionModel
	teleConeT::PredictionModel
	teleConeM::PredictionModel
	teleConeB::PredictionModel
	autoCubeT::PredictionModel
	autoCubeM::PredictionModel
	autoCubeB::PredictionModel
	teleCubeT::PredictionModel
	teleCubeM::PredictionModel
	teleCubeB::PredictionModel
	endgame::Dict{Int,PredictionModel}
end

function build_model23(gd::GameData)
    x = combine(
        filter(x -> length(x.key) == 3, groupby(gd.df, [:event, :key, :alliance,])),
        :team => (x -> [[teams(gd)[y] for y in x]]) => :teams,
		:auto_count_coneT => first => :auto_count_coneT,
		:auto_count_coneM => first => :auto_count_coneM,
		:auto_count_coneB => first => :auto_count_coneB,
		:teleop_count_coneT => first => :tele_count_coneT,
		:teleop_count_coneM => first => :tele_count_coneM,
		:teleop_count_coneB => first => :tele_count_coneB,
		:auto_count_cubeT => first => :auto_count_cubeT,
		:auto_count_cubeM => first => :auto_count_cubeM,
		:auto_count_cubeB => first => :auto_count_cubeB,
		:teleop_count_cubeT => first => :tele_count_cubeT,
		:teleop_count_cubeM => first => :tele_count_cubeM,
		:teleop_count_cubeB => first => :tele_count_cubeB,
    )

    team_auto_mobile = Dict{Int,PredictionModel}()
	for team in keys(teams(gd))
		team_auto_mobile[team] = team_auto_mobile_model(gd, team)
	end

	team_auto_charge = Dict{Int,PredictionModel}()
	for team in keys(teams(gd))
		team_auto_charge[team] = team_auto_charge_model23(gd, team)
	end

	endgame = Dict{Int,PredictionModel}()
    for team in keys(teams(gd))
        endgame[team] = team_endgame_model23(gd, team)
    end

	auto_coneT = run_model(gd, count_model(hcat(x.teams...), x.auto_count_coneT, length(gd.teams)))
	auto_coneM = run_model(gd, count_model(hcat(x.teams...), x.auto_count_coneM, length(gd.teams)))
	auto_coneB = run_model(gd, count_model(hcat(x.teams...), x.auto_count_coneB, length(gd.teams)))
	tele_coneT = run_model(gd, count_model(hcat(x.teams...), x.tele_count_coneT, length(gd.teams)))
	tele_coneM = run_model(gd, count_model(hcat(x.teams...), x.tele_count_coneM, length(gd.teams)))
	tele_coneB = run_model(gd, count_model(hcat(x.teams...), x.tele_count_coneB, length(gd.teams)))
	auto_cubeT = run_model(gd, count_model(hcat(x.teams...), x.auto_count_cubeT, length(gd.teams)))
	auto_cubeM = run_model(gd, count_model(hcat(x.teams...), x.auto_count_cubeM, length(gd.teams)))
	auto_cubeB = run_model(gd, count_model(hcat(x.teams...), x.auto_count_cubeB, length(gd.teams)))
	tele_cubeT = run_model(gd, count_model(hcat(x.teams...), x.tele_count_cubeT, length(gd.teams)))
	tele_cubeM = run_model(gd, count_model(hcat(x.teams...), x.tele_count_cubeM, length(gd.teams)))
	tele_cubeB = run_model(gd, count_model(hcat(x.teams...), x.tele_count_cubeB, length(gd.teams)))
    return Simulator23(
        gd, team_auto_mobile, team_auto_charge,
		auto_coneT, auto_coneM, auto_coneB,
		tele_coneT, tele_coneM, tele_coneB,
		auto_cubeT, auto_cubeM, auto_cubeB,
		tele_cubeT, tele_cubeM, tele_cubeB,
		endgame
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
			sample(model, NUTS(), 100; progress=false, verbose=false)
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

# TODO: This is almost certainly better with logit or similar.
@model function auto_charge(ye::Vector{Bool}, yd::Vector{Bool})
	e ~ Beta(1,1)
	d ~ Beta(1,1)
	ye .~ Bernoulli.(e)
	yd .~ Bernoulli.(d)
end

egda = Dict("Engaged"=>2, "Docked"=>1, "None"=>0)
function team_auto_charge_model23(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = auto_charge(
			[egda[y] >= 2 for y in df.auto_charge],
			[egda[y] >= 1 for y in df.auto_charge],
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 100; progress=false, verbose=false);
		end
	else
		model = auto_charge([true],[true])
		s = Logging.with_logger(logger) do
			sample(model, Prior(), 1000; progress=false)
		end
	end
	auchargedf = DataFrame(
		:team=>team,
		:engaged=>mean(collect(first(get(s, :e)))),
		:docked=>mean(collect(first(get(s, :d)))),
	)
	return PredictionModel(model, s, auchargedf)
end

function simulate_auto_charge_points(s::Chains, n::Int)
	levels = Dict(:e=>12, :d=>8)
	v = zeros(Int, 2, n)
	for (i, l) in enumerate([:e, :d])
		v[i,:] = levels[l] * rand.(Bernoulli.(rand(first(get(s, l)), n)))
	end
	return maximum(v, dims=1)[1, :]
end

function simulate_auto_charge_points(sim::Simulator23, team::Int, n::Int)
	chain = sim.endgame[team].chain
	return simulate_endgame_points(chain, n)
end

@model function endgame_model23(ye::Vector, yd::Vector, yp::Vector)
	e ~ Beta(1,1)
	d ~ Beta(1,1)
	p ~ Beta(1,1)
	ye .~ Bernoulli.(e)
	yd .~ Bernoulli.(d)
	yp .~ Bernoulli.(p)
	return e, d, p
end


egd = Dict("Engaged"=>4, "Docked"=>3, "Park"=>2, "None"=>1)
function team_endgame_model23(gd::GameData, team::Int)
	df = gd.df |> x -> x[x.team .== team, :]
	logger = Logging.NullLogger()
	if nrow(df) > 0
		model = endgame_model23(
			[egd[y] >= 4 for y in df.endgame_charge],
			[egd[y] >= 3 for y in df.endgame_charge],
			[egd[y] >= 2 for y in df.endgame_charge]
		)
		
		s = Logging.with_logger(logger) do
			sample(model, NUTS(), 100; progress=false, verbose=false);
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

function simulate_endgame_points(s::Chains, n::Int)
	levels = Dict(:e=>10, :d=>6, :p=>2)
	v = zeros(Int, 3, n)
	for (i, l) in enumerate([:e, :d, :p])
		v[i,:] = levels[l] * rand.(Bernoulli.(rand(first(get(s, l)), n)))
	end
	return maximum(v, dims=1)[1, :]
end

function simulate_endgame_points(sim::Simulator23, team::Int, n::Int)
	chain = sim.endgame[team].chain
	return simulate_endgame_points(chain, n)
end

function simulate_team_tuple(sim::Simulator23, team, n)
	auto_mobile = simulate_auto_mobile(sim.automobile[team].chain, n)
	auto_charge = simulate_auto_charge_points(sim, team, n)
	endgame = simulate_endgame_points(sim, team, n)
	return (
		auto_mobile=auto_mobile,
		auto_charge=auto_charge,
		auto_count_coneT=simulate_counts(sim.gd, sim.autoConeT, team, n),
		auto_count_coneM=simulate_counts(sim.gd, sim.autoConeM, team, n),
		auto_count_coneB=simulate_counts(sim.gd, sim.autoConeB, team, n),
		tele_count_coneT=simulate_counts(sim.gd, sim.teleConeT, team, n),
		tele_count_coneM=simulate_counts(sim.gd, sim.teleConeM, team, n),
		tele_count_coneB=simulate_counts(sim.gd, sim.teleConeB, team, n),
		auto_count_cubeT=simulate_counts(sim.gd, sim.autoCubeT, team, n),
		auto_count_cubeM=simulate_counts(sim.gd, sim.autoCubeM, team, n),
		auto_count_cubeB=simulate_counts(sim.gd, sim.autoCubeB, team, n),
		tele_count_cubeT=simulate_counts(sim.gd, sim.teleCubeT, team, n),
		tele_count_cubeM=simulate_counts(sim.gd, sim.teleCubeM, team, n),
		tele_count_cubeB=simulate_counts(sim.gd, sim.teleCubeB, team, n),
		endgame=endgame,
		other=0
	)
end

function simulate_team(sim::Simulator23, team, n)
	t = simulate_team_tuple(sim, team, n)
	return (
		3*t.auto_mobile .+
		t.auto_charge .+
		6*t.auto_count_coneT .+
		4*t.auto_count_coneM .+
		3*t.auto_count_coneB .+
		5*t.tele_count_coneT .+
		3*t.tele_count_coneM .+
		2*t.tele_count_coneB .+
		6*t.auto_count_cubeT .+
		4*t.auto_count_cubeM .+
		3*t.auto_count_cubeB .+
		5*t.tele_count_cubeT .+
		3*t.tele_count_cubeM .+
		2*t.tele_count_cubeB .+
		t.endgame
	)
end

function ev_team(sim::Simulator23, team::Int, n::Int)
	return simulate_team(sim, team, n)
end

function simulate_teams(sim::Simulator23, teams::Vector{Int}, n)
	return sum(
		[simulate_team(sim, x, n) for x in teams]
	)
end

function sim_evs(sim::Simulator23)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, 100_000) for y in x])
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