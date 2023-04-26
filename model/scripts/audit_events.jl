using DrWatson
@quickactivate

include(scriptsdir("frc2023.jl"))

using Turing
using ReverseDiff
using ThreadsX

Turing.setadbackend(:reversediff)
Turing.setrdcache(true)

want_events = ["2023arc", "2023cur", "2023dal", "2023gal", "2023hop", "2023joh", "2023mil", "2023new"]
a = ThreadsX.map(produce_audit, want_events)
GC.gc()