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
         x -> sort(x, :time) #|> x -> x[x.comp_level .== "qm", :] |> x -> x[x.match_number .<= 8, :]

function ev_count_df(sim::FRCModels.Simulator23)
	return FRCModels.sim_evs(sim) |> 
		x -> DataFrame(:team=>x.teams, :points => x.sims) |>
		x -> flatten(x, :points) |>
		x -> combine(groupby(x, [:team, :points]), nrow=>:count)
end

function save_event_data(event)
	df = df_all |> x -> x[x.event .== event, :]
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
	#pred = FRCModels.bymatch(df) |> x -> FRCModels.win_probabilities.(Ref(mod), x.red, x.blue)
	#match_data.predictions = first.(pred)
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
	open("../files/api/events/$(event).json","w") do f
		write(f, out)
	end
end

for event in Set(["2023week0"])
	println(event)
	save_event_data(event)
end
