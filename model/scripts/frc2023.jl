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
import StatsBase: countmap, mean, std

###
df_all = dropmissing(DataFrame(Arrow.Table(datadir("raw", "frc2023.feather")))) |>
         x -> sort(x, :time) #|> x -> x[x.comp_level .== "qm", :] |> x -> x[x.match_number .<= 15, :]

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

function get_schedule(event)
	schedule = dropmissing(DataFrame(Arrow.Table(datadir("schedules", "$(event).feather"))))
	schedule.red_teams = map(x -> [y for y in x if !ismissing(y)], schedule.red_teams)
	schedule.blue_teams = map(x -> [y for y in x if !ismissing(y)], schedule.blue_teams)
	return schedule[(0 .∉ schedule.red_teams) .& (0 .∉ schedule.blue_teams), :]
end

function save_event_data(event; time=2077617135)
	df = df_all |> x -> x[x.event .== event, :]
	dd = df |> x -> x[x.time .< time, :]
	#|> x -> x[x.comp_level .== "qm", :] |> x -> x[x.match_number .<= 20, :]
	gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
	sim = FRCModels.build_model23(gd)
	y = sim |>
			x -> rename(ev_count_df(x), :count=>:bcount) |>
			x -> transform(x, :team=>ByRow(string)=>:team)
	ev = groupby(y, :team) |> 
		collect .|> 
		x -> Dict("points"=>x.points, "bcount"=>x.bcount, "team"=>first(x.team))
	match_data = df |> FRCModels.bymatch
	team_simulations = Dict{Int, Dict}()
	for team in sim.gd.teams
		sim_data = FRCModels.simulate_team_tuple(sim, team, 1_000)
		di = Dict()
		for key in keys(sim_data)
			di[key] = countmap(sim_data[key])
		end
		team_simulations[team] = di
	end
	schedule = get_schedule(event)
	# for testing
	#=
	schedule[schedule.comp_level .!== "qm", :red_score] .= -1
	schedule[schedule.comp_level .!== "qm", :blue_score] .= -1
	schedule[(schedule.comp_level .== "qm") .& (schedule.match_number .> 20), :red_score] .= -1
	schedule[(schedule.comp_level .== "qm") .& (schedule.match_number .> 20), :blue_score] .= -1
	=#
	#
	predictions = build_predictions(sim, schedule)
	return ev, match_data, team_simulations, predictions, schedule
	
end

function write_event(event, ev, match_data, team_simulations, predictions, schedule)
	out = "{\"ev\":$(JSON3.write(ev)),\"matches\":$(arraytable(match_data)),\"team_sims\":$(JSON3.write(team_simulations)),\"predictions\":$(JSON3.write(predictions)),\"schedule\":$(arraytable(schedule))}"
	open("../files/api/events/$(event).json","w") do f
		write(f, out)
	end
end

function audit_event(event, df_all)
	df = df_all |> x -> x[x.event .== event, :] |> FRCModels.bymatch
	predictions = Dict{String,Dict{String,Number}}()
	for (i, r) in enumerate(eachrow(df))
		if i <= 1 #|| i > 30
			continue
		end
		println(r["key"])
		df_e = df_all |> x -> x[x.event .== event, :] #|> x -> x[x.time .< time, :]
		dd = df_e |> x -> x[x.time .< r["time"], :]
		gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
		sim = FRCModels.build_model23(gd)
		#sched = get_schedule(event)
		redsim, bluesim = r |> x -> FRCModels.simulate_match(sim, x.red, x.blue; n=10_000)
		pred = build_predictions(sim, redsim, bluesim)
		predictions[r["key"]] = pred
	end
	return predictions
end

#=
event = "2023isde1"
@time pred = audit_event(event, df_all[(df_all.event .== event), :])
pdf = DataFrame(values(pred))
pdf[:, "key"] = collect(keys(pred))
sched = get_schedule(event)
rdf = leftjoin(pdf, sched, on=:key)
rdf[:, "brier"] = ((rdf.red_score .> rdf.blue_score) .- rdf.red_win).^2
=#

for event in Set(["2023bcvi", "2023isde2", "2023isde1"])
	println(event)
	ev, match_data, team_simulations, predictions, sched = save_event_data(event)
	write_event(event, ev, match_data, team_simulations, predictions, sched)
end
