module FRCModels

using DataFrames
using Arrow
using CSV
using ReverseDiff
using LinearAlgebra
using Logging
using LazyArrays
using Turing
import OrderedCollections: OrderedSet

struct GameData
	df::DataFrame
	teams::OrderedSet{Int}
end

struct PredictionModel
	model::Turing.DynamicPPL.Model
	chain::Turing.Chains
	summary::DataFrame
end

function bymatch(df::DataFrame)
	combine(
        filter(y -> length(y.key) == 6 && sum([y.score[1], y.score[6]]) != 1, groupby(df, [:key])),
        :team => (y -> [y[1:3]]) => :red,
        :team => (y -> [y[4:6]]) => :blue,
        :score => (y -> y[1]) => :red_score,
        :score => (y -> y[6]) => :blue_score,
        :winner => first => :winner,
        :match_number => first => :match_number,
        :time => first => :time,
    )
end

"""
    samples_df(s, teamsdict)

    Create a summary DataFrame from chain data.
"""
function samples_df(s::Turing.Chains, teamsdict::Dict{Int, Int})
	off = collect(first(get(s, :off)))
	return DataFrame(
		:team=>[teamsdict[x] for x in 1:length(off)],
		:mean=>mean.(off),
		:std=>std.(off),
	)
end

function teams(gd::GameData)
	return  gd.teams |> x -> Dict(x.=>1:length(x))
end

function teamsr(gd::GameData)
	return  gd.teams |> x -> Dict(1:length(x).=>x)
end

"""
    _q(v::Vector, n::Int)

    Get the `n`th element in `v`.
"""
function _q(v::Vector, n::Int)
	return sort(v)[n]
end

function run_model(gd::GameData, model; fast=false)
	if fast
		sampler = HMC(0.05, 1)
		samples = 1000
	else
		sampler = NUTS()
		samples = 100
	end
	logger = Logging.NullLogger()
	s = Logging.with_logger(logger) do
		sample(model, sampler, samples; drop_warmup=true, progress=false, verbose=false)
	end
	summary = samples_df(s, teamsr(gd))
	sort!(summary, :mean; rev=true)
	return PredictionModel(model, s, summary)
end

include("models.jl")
include("models22.jl")
include("simulator22.jl")
include("simulator23.jl")

end # module FRCModels
