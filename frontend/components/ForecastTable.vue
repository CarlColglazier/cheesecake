<template>
	<div>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Team</b-th>
					<b-th>Firest Seed?</b-th>
					<b-th>Captain? (TODO)</b-th>
				</b-tr>
			</b-thead>
			<b-tbody>
				<b-tr v-for="team in allTeams(forecasts)" :key="team">
					<b-td>{{ team }}</b-td>
					<b-td>
						<svg width="150" height="50">
							<line v-for="cast in teamForecasts(forecasts, team)"
													 :x1="cast.match" :x2="cast.match"
													 y2="50" :y1="50 - cast.forecast * 50"
													 stroke="black" stroke-width="5"
							/>
						</svg>
						<span>{{ latestForecast(forecasts, team) }}</span>
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
	var max_match = Math.max.apply(Math, forecasts.map(function(f) { return f.match }));
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

function topTeams(forecast) {
	var max_match = Math.max.apply(Math, forecast.map(function(f) { return f.match }));
	var max_fore = forecast.filter(f => {
		return f.match === max_match;
	});
	return max_fore.map(function(f) { return f.team });
}


export default {
	methods: {
		allTeams: allTeams,
		latestForecast: latestForecast,
		teamForecasts: teamForecasts
	},
	props: ['forecasts']
}
</script>
