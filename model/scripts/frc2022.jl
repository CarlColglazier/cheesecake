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

###
df_all = dropmissing(DataFrame(Arrow.Table(datadir("raw", "frc2022.feather")))) |>
         x -> sort(x, :time)

elos = CSV.read(datadir("raw", "elo.csv"), DataFrame) |>
	x -> Dict(x.Team .=> x.Elo)

x = FRCModels.run_event_once(df_all, "2022nccmp", elos);

# TODO: Saving and loading model doesn't produce the same results!


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

#=
for event in Set(df_all.event)
	println(event)
	x = FRCModels.Simulator(load(datadir("simulations", "$(event).jld2")))
	df = ev_count_df(x)
	Arrow.write(datadir("out", "ev", "$(event).feather"), df)
end
=#

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