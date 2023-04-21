using DrWatson
@quickactivate

using Revise # hot reload

using FRCModels
using Arrow
using DataFrames
using CSV
import OrderedCollections: OrderedSet

using JSON3
using JSONTables
import StatsBase: countmap, mean, std, median

elos = CSV.read(datadir("raw", "elo.csv"), DataFrame) |>
	x -> Dict(x.Team .=> x.Elo)
elo_t = CSV.read(datadir("raw", "elo.csv"), DataFrame) |>
	x -> Dict(x.Team .=> (x.Elo .- 1450)./1000)


function ev_count_df(sim::FRCModels.Simulator23)
	return FRCModels.sim_evs(sim) |> 
		x -> DataFrame(:team=>x.teams, :points => x.sims) |>
		x -> flatten(x, :points) |>
		x -> combine(groupby(x, [:team, :points]), nrow=>:count)
end

function build_predictions(sim::FRCModels.Simulator23, redsim, bluesim; n=10_000)
	redpoints = FRCModels.simulate_teams(redsim)
	bluepoints = FRCModels.simulate_teams(bluesim)
	pred = [
		sum(bluepoints .< redpoints) / n,
		sum(bluepoints .== redpoints) / n,
		sum(bluepoints .> redpoints) / n,
	]
	d = Dict{String,Number}()
	d["red_win"] = pred[1]
	d["blue_win"] = pred[3]
	d["tie"] = pred[2]
	d["red_activation"] = mean(redsim.activation)
	d["blue_activation"] = mean(bluesim.activation)
	d["red_sustainability"] = mean(redsim.sustainability)
	d["blue_sustainability"] = mean(bluesim.sustainability)
	d["red_score_median"] = median(FRCModels.score(redsim))
	d["blue_score_median"] = median(FRCModels.score(bluesim))
	return d
end

function build_predictions(sim::FRCModels.Simulator23, matches::DataFrame; n=10_000)
	predictions = Dict{String,Dict{String,Number}}()
	for r in eachrow(matches)
		redsim, bluesim = r |> x -> FRCModels.simulate_match(sim, x.red_teams, x.blue_teams; n=n)
		predictions[r["key"]] = build_predictions(sim, redsim, bluesim)
	end
	return predictions
end

function get_schedule(event::String)
	schedule = dropmissing(DataFrame(Arrow.Table(datadir("schedules", "$(event).feather"))))
	schedule.red_teams = map(x -> [y for y in x if !ismissing(y)], schedule.red_teams)
	schedule.blue_teams = map(x -> [y for y in x if !ismissing(y)], schedule.blue_teams)
	return schedule[(0 .∉ schedule.red_teams) .& (0 .∉ schedule.blue_teams), :]
end

function get_breakdown(event::String)
	df = dropmissing(DataFrame(Arrow.Table(datadir("breakdowns", "$(event).feather"))))
	return sort(df, :time)
end

function save_event_data(event::String)
	df = get_breakdown(event) #|> x -> x[(x.comp_level .== "qm") .& (x.match_number .<= 12), :]
	#|> x -> x[(x.comp_level .== "qm") .| ((x.comp_level .== "sf") .& (x.match_number .< 10)), :]
	schedule = get_schedule(event)
	sched_teams = Set(collect(Iterators.flatten(reduce(vcat, [schedule.red_teams, schedule.blue_teams]))))
	gd = FRCModels.GameData(df, sched_teams)
	println("Building simulation for $(event)")
	sim = FRCModels.build_model23(gd, elo_t)
	println("Sumulating EV for $(event)")
	y = sim |>
			x -> rename(ev_count_df(x), :count=>:bcount) |>
			x -> transform(x, :team=>ByRow(string)=>:team)
	ev = groupby(y, :team) |> 
		collect .|> 
		x -> Dict("points"=>x.points, "bcount"=>x.bcount, "team"=>first(x.team))
	match_data = df |> FRCModels.bymatch23
	println("Team simulations for $(event)")
	team_simulations = Dict{Int, Dict}()
	for team in sim.gd.teams
		sim_data = FRCModels.simulate_team_tuple(sim, team, 1_000)
		di = Dict()
		for key in keys(sim_data)
			di[key] = countmap(sim_data[key])
		end
		team_simulations[team] = di
	end
	println("Predictions for $(event)")
	predictions = build_predictions(sim, schedule)
	return sim, ev, match_data, team_simulations, predictions, schedule
end

function model_summary(sim::FRCModels.Simulator23)
	#model_keys = [:autoT, :autoM, :autoB, :teleT, :teleM, :teleB]
	model_keys = [:autoT, :teleT]
	return join(map(x -> "\"" * string(x) * "\":" * arraytable(getproperty(sim, x).summary), model_keys), ",")
end

function write_event(sim::FRCModels.Simulator23, event::String, ev, match_data, team_simulations, predictions, schedule)
	#tagsave(datadir("simulations", "$(event).jld2"), struct2dict(sim))
	out = "{\"ev\":$(JSON3.write(ev)),\"matches\":$(arraytable(match_data)),\"team_sims\":$(JSON3.write(team_simulations)),\"predictions\":$(JSON3.write(predictions)),\"schedule\":$(arraytable(schedule)),\"model_summary\":{$(model_summary(sim))}}"
	open("../files/api/events/$(event).json","w") do f
		write(f, out)
	end
end

function audit_event_match(event, df_all, i, elo_t)
	df = df_all |> x -> x[x.event .== event, :] |> FRCModels.bymatch
	r = df[i, :]
	df_e = df_all |> x -> x[x.event .== event, :] #|> x -> x[x.time .< time, :]
	dd = df_e |> x -> x[x.time .< r["time"], :]
	gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
	sim = FRCModels.build_model23(gd, elo_t)
	redsim, bluesim = r |> x -> FRCModels.simulate_match(sim, x.red, x.blue; n=10_000)
	pred = build_predictions(sim, redsim, bluesim)
	return pred
end

function audit_event(event, df_all)
	df = df_all |> x -> x[x.event .== event, :] |> FRCModels.bymatch
	predictions = Dict{String,Dict{String,Number}}()
	for (i, r) in enumerate(eachrow(df))
		if i <= 1
			continue
		end
		println(r["key"])
		df_e = df_all |> x -> x[x.event .== event, :] #|> x -> x[x.time .< time, :]
		dd = df_e |> x -> x[x.time .< r["time"], :]
		gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
		sim = FRCModels.build_model23(gd, elo_t)
		redsim, bluesim = r |> x -> FRCModels.simulate_match(sim, x.red, x.blue; n=10_000)
		pred = build_predictions(sim, redsim, bluesim)
		println(pred["red_win"])
		predictions[r["key"]] = pred
	end
	return predictions
end

function produce_audit(event)
	df_all = sort(get_breakdown(event), :time)
	@time pred = audit_event(event, df_all[(df_all.event .== event), :]);
	pdf = DataFrame(values(pred))
	pdf[:, "key"] = collect(keys(pred))
	sched = get_schedule(event)
	rdf = leftjoin(pdf, sched, on=:key)
	rdf[:, "brier"] = ((rdf.red_score .> rdf.blue_score) .- rdf.red_win).^2
	return rdf
end

function event_wants_update(event)
	return mtime("../files/api/events/$(event).json") < mtime(datadir("breakdowns", "$(event).feather"))
end

function list_events()
	return first.(splitext.(readdir(datadir("breakdowns"))))
end

function save_events(events)
	open("../files/api/event_keys.json", "w") do f
		write(f, JSON3.write(events))
	end
end

function run_event(event)
	if !event_wants_update(event)
		return
	end
	println(event)
	try
		sim, ev, match_data, team_simulations, predictions, sched = save_event_data(event)
		if !isdir("../files/api/events/")
			mkdir("../files/api/events/")
		end
		write_event(sim, event, ev, match_data, team_simulations, predictions, sched)
	catch e
		println(e)
		stacktrace(e)
		return
	end
	GC.safepoint()
end


