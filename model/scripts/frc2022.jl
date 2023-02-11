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
import StatsBase: countmap

###
df_all = dropmissing(DataFrame(Arrow.Table(datadir("raw", "frc2022.feather")))) |>
         x -> sort(x, :time)

elos = CSV.read(datadir("raw", "elo.csv"), DataFrame) |>
	x -> Dict(x.Team .=> x.Elo)

x = FRCModels.run_event_once(df_all, "2022nccmp", elos);

for event in Set(df_all.event)
	println(event)
	x = FRCModels.run_event_once(df_all, event, elos);
	tagsave(datadir("simulations", "$(event).jld2"), struct2dict(x))
end

function ev_count_df(sim::FRCModels.Simulator)
	return FRCModels.sim_evs(sim) |> 
		x -> DataFrame(:team=>x.teams, :points => x.sims) |>
		x -> flatten(x, :points) |>
		x -> combine(groupby(x, [:team, :points]), nrow=>:count)
end

open("files/events.json", "w") do f 
	JSON.print(f, Set(df_all.event))
end

for event in Set(df_all.event)
	println(event)
	x = FRCModels.run_event_once(df_all, event, elos);
	y = x |>
		x -> rename(ev_count_df(x), :count=>:bcount) |>
		x -> transform(x, :team=>ByRow(string)=>:team)
	d = groupby(y, :team) |> 
		collect .|> 
		x -> Dict("points"=>x.points, "bcount"=>x.bcount, "team"=>first(x.team))
	match_data = x.gd.df |> FRCModels.bymatch
	predictions = FRCModels.run_event(df_all, event, elos)
	match_data.predictions = predictions
	out = "{\"ev\":$(JSON3.write(d)),\"matches\":$(arraytable(match_data))}"
	open("files/api/events/$(event).json","w") do f
		write(f, out)
		#JSON.print(f, out)
	end
end

function save_event_data(event)
	df = df_all |> x -> x[x.event .== event, :] #x[(x.event .== event) .& (x.comp_level .== "qm"), :]
	#df = df_all |> x -> x[(x.event .== event), :]
	#Dict(Set(df.team) .=> 1500)
	dd = df#df[df.match_number .<= 100, :]
	gd = FRCModels.GameData(dd, Set(df_all[df_all.event .== event, :team]))
	mod = FRCModels.build_model(gd, elos)
	#mod = FRCModels.run_event_once(df[df.match_number .<= 1, :], event, Dict(Set(df.team) .=> 1500));
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
	out = "{\"ev\":$(JSON3.write(d)),\"matches\":$(arraytable(match_data)),\"team_sims\":$(JSON3.write(team_simulations))}"
	open("files/api/events/$(event).json","w") do f
		write(f, out)
	end
end

# fake some data for testing
event = "2022nccmp"
for event in Set(df_all.event)
	println(event)
	save_event_data(event)
end
