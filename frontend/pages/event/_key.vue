<template>
	<content>
		<table>
			<thead>
				<tr>
					<th>Key</th>
					<th colspan="1">
						Red
					</th>
					<th colspan="1">
						Blue
					</th>
					<th colspan="2">
						Score
					</th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="d in matches" :key="d.match.key">
					<td>
						{{
							d.match.comp_level +
								(d.match.comp_level != 'qm' ? d.match.set_number : '') +
								d.match.match_number
						}}
					</td>
					<td>
						<span
							v-for="team in d.alliances.red.teams"
							:key="team"
							class="break"
						>
							{{ team.substring(3) }}
						</span>
					</td>
					<td>
						<span
							v-for="team in d.alliances.blue.teams"
							:key="team"
							class="break"
						>
							{{ team.substring(3) }}
						</span>
					</td>
					<td>
						<span class="break">
							{{ d.alliances.red.alliance.score }}
						</span>
						<span class="break">
							{{ displayPrediction(d.predictions.elo_score.prediction, 'red') }}
						</span>
					</td>
					<td>
						<span class="break">
							{{ d.alliances.blue.alliance.score }}
						</span>
						<span class="break">
							{{
								displayPrediction(d.predictions.elo_score.prediction, 'blue')
							}}
						</span>
					</td>
				</tr>
			</tbody>
		</table>
	</content>
</template>

<style>
span.break {
	display: block;
}

td, th {
    text-align: center;
    padding: 1em;
}
</style>

<script>
let url = (process.server) ? 'http://backend:8080' : 'http://localhost:8080';

function rankPoints(input) {
	if (input) {
		return '\u2713'
	}
	return ''
}

function prediction(num, color) {
	if (num === null) {
		return ''
	}
	if (color === 'blue') {
		num = 1 - num
	}
	const percent = Math.round((num - 0.5) * 100)
	if (percent > 0) {
		return `(${Math.round(num * 100)}%)`
	}
	return ''
}

export default {
	async asyncData({ params }) {
		const dataf = await fetch(url + '/matches/' + params.key)
		const data = await dataf.json()
		return { matches: data }
	},
	methods: {
		rankPoints: rankPoints,
		displayPrediction: prediction
	}
}
</script>
