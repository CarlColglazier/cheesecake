<template>
	<div>
		<p>Forecast current as of Match {{ maxMatch(forecasts.rpleader) }}</p>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Team</b-th>
					<b-th>Graphic</b-th>
					<b-th>First Seed?</b-th>
					<b-th>Captain?</b-th>
					<b-th>Mean RP</b-th>
				</b-tr>
			</b-thead>
			<b-tbody>
				<b-tr v-for="team in sortedTeams" :key="team">
					<b-td>{{ team }}</b-td>
					<b-td>
						<svg width="100" height="50">
							<line v-for="cast in teamForecasts(forecasts.cap, team)"
													 :x1="cast.match" :x2="cast.match"
													 y2="50" :y1="50 - cast.forecast * 50"
													 stroke="#999" stroke-width="1"
							/>
							<line v-for="cast in teamForecasts(forecasts.rpleader, team)"
													 :x1="cast.match" :x2="cast.match"
													 y2="50" :y1="50 - cast.forecast * 50"
													 stroke="#000" stroke-width="1"
							/>
						</svg>
					</b-td>
					<b-td>{{ latestForecast(forecasts.rpleader, team) }}</b-td>
					</b-td>
					<b-td>
						{{ latestForecast(forecasts.cap, team) }}
					</b-td>
					<b-td>
						<svg width="100" height="50">
							<line v-for="cast in teamForecasts(forecasts.meanrp, team)"
										:x1="cast.match" :x2="cast.match"
													 :y2="50 - cast.forecast + 3.92 * getStd(forecasts, team, cast.match)"
													 :y1="50 - cast.forecast - 3.92 * getStd(forecasts, team, cast.match)"
										stroke="#888" stroke-width="1"
							/>
							<line v-for="cast in teamForecasts(forecasts.meanrp, team)"
													 :x1="cast.match" :x2="cast.match"
													 :y2="50 - cast.forecast + 1"
													 :y1="50 - cast.forecast - 1"
													 stroke="black" stroke-width="1"
							/>
							<!--
									 getStd(forecasts, team, cast.match)
									 <line v-for="cast in teamForecasts(forecasts.stdrp, team)"
									 :x1="cast.match" :x2="cast.match"
									 y2="30" :y1="30 - cast.forecast * 3.92"
									 stroke="#999" stroke-width="1"
									 />
							-->
						</svg>
						<span>{{ Math.round(latestForecast(forecasts.meanrp, team)) }}</span>
					</b-td>
				</b-tr>
			</b-tbody>
		</b-table-simple>
	</div>
</template>

<script>
function allTeams(forecasts) {
	return forecasts.map(m => {
		return m.team;
	}).filter((v, i, a) => {
		return a.indexOf(v) === i;
	}); 
}

function latestForecast(forecasts, team) {
	var max_match = maxMatch(forecasts);
	var max_fore = forecasts.filter(f => {
		return f.match === max_match && f.team === team;
	});
	if (max_fore.length < 1) {
		return "<1%";
	}
	return max_fore[0].forecast;
}

function teamForecasts(forecasts, team) {
	return forecasts.filter(f => {
		return f.team === team;
	});
}

function maxMatch(forecast) {
	return Math.max.apply(Math, forecast.map(function(f) { return f.match }));
}

function topTeams(forecast) {
	var max_match = maxMatch(forecast);
	var max_fore = forecast.filter(f => {
		return f.match === max_match;
	});
	return max_fore.map(function(f) { return f.team });
}

function getStd(forecasts, team, match) {
	var fore = forecasts.stdrp.filter(f => {
		return f.match === match && f.team === team;
	});
	return fore[0].forecast;
}

function sortedTeams() {
	let forecasts = this.forecasts;
	let teams = allTeams(forecasts.meanrp).sort((a, b) => {
		return latestForecast(forecasts.meanrp, a) - latestForecast(forecasts.meanrp, b);
	}).reverse();
	return teams;
}


export default {
	methods: {
		allTeams: allTeams,
		latestForecast: latestForecast,
		teamForecasts: teamForecasts,
		maxMatch: maxMatch,
		getStd: getStd
	},
	computed: {
		sortedTeams: sortedTeams
	},
	props: ['forecasts']
}
</script>
