<template>
	<content>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Key</b-th>
					<b-th colspan="1">
						Red
					</b-th>
					<b-th colspan="1">
						Blue
					</b-th>
					<b-th colspan="2">
						Score
					</b-th>
				</b-tr>
			</b-thead>
			<b-tbody>
				<b-tr v-for="d in matches" :key="d.match.key">
					<b-td>
						{{
							d.match.comp_level +
								(d.match.comp_level != 'qm' ? d.match.set_number : '') +
								d.match.match_number
						}}
					</b-td>
					<b-td>
						<span
							v-for="team in d.alliances.red.teams"
							:key="team"
							class="break"
						>
							{{ team.substring(3) }}
						</span>
					</b-td>
					<b-td>
						<span
							v-for="team in d.alliances.blue.teams"
							:key="team"
							class="break"
						>
							{{ team.substring(3) }}
						</span>
					</b-td>
					<b-td>
						<span class="break">
							{{ d.alliances.red.alliance.score }}
						</span>
						<span class="break">
							{{ displayPrediction(d.predictions.elo_score.prediction, 'red') }}
						</span>
					</b-td>
					<b-td>
						<span class="break">
							{{ d.alliances.blue.alliance.score }}
						</span>
						<span class="break">
							{{
								displayPrediction(d.predictions.elo_score.prediction, 'blue')
							}}
						</span>
					</b-td>
				</b-tr>
			</b-tbody>
		</b-table-simple>
	</content>
</template>

<style>
span.break {
	display: block;
}
</style>

<script>
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
	layout: 'default',
	async asyncData(context) {
		try {
			const dataf = await context.app.fetch(`/matches/${context.params.key}`)
			const data = await dataf.json()
			return { matches: data }
		} catch (e) {
			console.error(e)
			return { matches: [] }
		}
	},
	methods: {
		rankPoints: rankPoints,
		displayPrediction: prediction
	}
}
</script>
