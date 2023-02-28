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

function build_predictions(sim::FRCModels.Simulator23, matches)
	n = 10_000
	predictions = Dict{String,Dict{String,Number}}()
	for (i, r) in enumerate(eachrow(matches))
		redsim, bluesim = r |> x -> FRCModels.simulate_match(sim, x.red, x.blue; n=n)
		redpoints = FRCModels.simulate_teams(redsim)
		bluepoints = FRCModels.simulate_teams(bluesim)
		pred = [
			sum(bluepoints .< redpoints) / n,
			sum(bluepoints .== redpoints) / n,
			sum(bluepoints .> redpoints) / n,
		]
		d = Dict{String,Number}()
		#d["key"] = r["key"]
		d["red_win"] = pred[1]
		d["blue_win"] = pred[3]
		d["tie"] = pred[2]
		d["red_activation"] = mean(redsim.activation)
		d["blue_activation"] = mean(bluesim.activation)
		d["red_sustainability"] = mean(redsim.sustainability)
		d["blue_sustainability"] = mean(bluesim.sustainability)
		#push!(predictions, d)
		predictions[r["key"]] = d
	end
	return predictions
end

function save_event_data(event)
	df = df_all |> x -> x[x.event .== event, :]
	#df.auto_countT = df.auto_count_coneT .+ df.auto_count_cubeT
	#df.teleop_countT = df.teleop_count_coneT .+ df.teleop_count_cubeT
	dd = df#df[df.match_number .<= 100, :]
	gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
	mod = FRCModels.build_model23(gd)
	y = mod |>
			x -> rename(ev_count_df(x), :count=>:bcount) |>
			x -> transform(x, :team=>ByRow(string)=>:team)
	d = groupby(y, :team) |> 
		collect .|> 
		x -> Dict("points"=>x.points, "bcount"=>x.bcount, "team"=>first(x.team))
	match_data = df |> FRCModels.bymatch
	pred = FRCModels.bymatch(df) |> x -> FRCModels.win_probabilities.(Ref(mod), x.red, x.blue)
	match_data.predictions = first.(pred)
	team_simulations = Dict{Int, Dict}()
	for team in mod.gd.teams
		sim_data = FRCModels.simulate_team_tuple(mod, team, 1_000)
		di = Dict()
		for key in keys(sim_data)
			di[key] = countmap(sim_data[key])
		end
		team_simulations[team] = di
	end
	predictions = build_predictions(mod, FRCModels.bymatch(df))
	out = "{\"ev\":$(JSON3.write(d)),\"matches\":$(arraytable(match_data)),\"team_sims\":$(JSON3.write(team_simulations)),\"predictions\":$(JSON3.write(predictions))}"
	open("../files/api/events/$(event).json","w") do f
		write(f, out)
	end
end

for event in Set(["2023week0", "2023isde1"])
	println(event)
	save_event_data(event)
end
