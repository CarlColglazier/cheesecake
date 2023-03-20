using DrWatson
@quickactivate

include(scriptsdir("frc2023.jl"))

all_events = collect(JSON3.read(read("../files/api/events.json", String)))
want_events = filter(event_wants_update, all_events)
n_threads = Threads.nthreads()
@sync for i in 1:n_threads
	Threads.@spawn for j in i:n_threads:length(want_events)
		run_event(want_events[j])
	end
end