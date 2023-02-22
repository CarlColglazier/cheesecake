struct Simulator22
	gd::GameData
	taxi::Dict{Int,PredictionModel}
	cargoupper::PredictionModel
	cargolower::PredictionModel
	cargoautoupper::PredictionModel
	cargoautolower::PredictionModel
	climb::Dict{Int, PredictionModel}
	fouls::PredictionModel
end

Simulator22(d::Dict{String, Any}) = Simulator22(
    d["gd"], d["cargoupper"], d["cargolower"], d["cargoautoupper"], d["cargoautolower"], d["climb"], d["fouls"]
)

function build_model22(gd::GameData, elos)
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
    return Simulator22(
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

function simulate_climb_points(sim::Simulator22, team::Int, n::Int)
	chain = sim.climb[team].chain
	return simulate_climb_points(chain, n)
end

function simulate_count_teams(gd::GameData, pm::PredictionModel, teams::Vector{Int}, n::Int)
	return sum(simulate_counts.(Ref(gd), Ref(pm), teams, n))
end

function simulate_total_cargo(sim::Simulator22, teams::Vector{Int}, n::Int)
	return (
		simulate_count_teams(sim.gd, sim.cargoautolower, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargoautoupper, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargoupper, teams, n) .+
		simulate_count_teams(sim.gd, sim.cargolower, teams, n)
	)
end

function simulate_team_tuple(sim::Simulator22, team, n)
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

function simulate_team(sim::Simulator22, team, n)
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

function ev_team(sim::Simulator22, team::Int, n::Int)
	# TODO: Fouls
	return simulate_team(sim, team, n) #.- 4*simulate_counts(sim.gd, sim.fouls, team, n)
end

function simulate_teams(sim::Simulator22, teams::Vector{Int}, n)
	return sum(
		[simulate_team(sim, x, n) for x in teams]
	)
end

function win_probabilities(sim::Simulator22, blue::Vector{Int}, red::Vector{Int}; n = 10_000)
	bluesim = simulate_teams(sim, blue, n) .+ 4*simulate_count_teams(sim.gd, sim.fouls, red, n)
	redsim = simulate_teams(sim, red, n) .+ 4*simulate_count_teams(sim.gd, sim.fouls, blue, n)
	return [
		sum(bluesim .> redsim) / n,
		sum(bluesim .== redsim) / n,
		sum(bluesim .< redsim) / n,
	]
end

function run_event_once22(df, key, elos)
	event_matches = df |>
		x -> x[x.event .== key, :] |> 
		x -> sort(x, :time)
	teams = OrderedSet(sort(event_matches.team))
	x = bymatch(event_matches)
	gddf = event_matches # completed?
	gd = GameData(gddf, teams)
	sim = build_model22(gd, elos)
	return sim
end

#### Run the model for an entire event!
function run_event22(df, key, elos)
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
		sim = build_model22(gd, elos);
		p = win_probabilities(sim, r.red, r.blue)
		push!(predictions, first(p))
		#push!(predictions, [r.key, first(p), r.winner .== "red"])
		#push!(simulations, sim)
	end
	return predictions#, simulations
end

function sim_evs(sim::Simulator22)
	return collect(sim.gd.teams) |>
	    x -> (teams=x, sims=[ev_team(sim, y, 100_000) for y in x])
end

function evs(sim::Simulator22)
	return sim_evs(sim) |>
	    x -> DataFrame(
		    :team => x[:teams],
		    :ev_median => median.(x[:sims]),
		    :ev_low => _q.(x[:sims], Int(100_000*0.05)),
            :ev_high => _q.(x[:sims], Int(100_000*0.95))
		) |>
		x -> sort(x, :ev_median; rev=true)
end