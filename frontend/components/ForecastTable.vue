<template>
	<div>
		<p>Forecast current as of Match {{ maxMatch(forecasts.rpleader) }}</p>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Team</b-th>
					<b-th>First Seed?</b-th>
					<b-th>Captain?</b-th>
				</b-tr>
			</b-thead>
			<b-tbody>
				<b-tr v-for="team in allTeams(forecasts.cap).sort(function(a, b) { return a < b }).reverse()" :key="team">
					<b-td>{{ team }}</b-td>
					<b-td>
						<svg width="150" height="30">
							<line v-for="cast in teamForecasts(forecasts.rpleader, team)"
										:x1="cast.match" :x2="cast.match"
										y2="30" :y1="30 - cast.forecast * 30"
										stroke="black" stroke-width="1"
							/>
						</svg>
						<span>{{ latestForecast(forecasts.rpleader, team) }}</span>
					</b-td>
					<b-td>
						<svg width="130" height="30">
							<line v-for="cast in teamForecasts(forecasts.cap, team)"
													 :x1="cast.match" :x2="cast.match"
													 y2="30" :y1="30 - cast.forecast * 30"
													 stroke="black" stroke-width="1"
							/>
						</svg>
						<span>{{ latestForecast(forecasts.cap, team) }}</span>
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


export default {
	methods: {
		allTeams: allTeams,
		latestForecast: latestForecast,
		teamForecasts: teamForecasts,
		maxMatch: maxMatch
	},
	props: ['forecasts']
}
</script>
