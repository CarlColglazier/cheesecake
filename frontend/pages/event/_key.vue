<template>
	<div>
		<v-simple-table>
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
					<th colspan="2">
						Rocket
					</th>
					<th colspan="2">
						Hab
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
							{{ displayPrediction(d.predictions.EloScore) }}
						</span>
					</td>
					<td>
						<span class="break">
							{{ d.alliances.blue.alliance.score }}
						</span>
						<span class="break">
							{{ displayPrediction(1 - d.predictions.EloScore) }}
						</span>
					</td>
					<td>
						{{
							rankPoints(d.match.score_breakdown.red.completeRocketRankingPoint)
						}}
					</td>
					<td>
						{{
							rankPoints(
								d.match.score_breakdown.blue.completeRocketRankingPoint
							)
						}}
					</td>
					<td>
						{{ rankPoints(d.match.score_breakdown.red.habDockingRankingPoint) }}
					</td>
					<td>
						{{
							rankPoints(d.match.score_breakdown.blue.habDockingRankingPoint)
						}}
					</td>
				</tr>
			</tbody>
		</v-simple-table>
	</div>
</template>

<style>
span.break {
	display: block;
}
</style>

<script>
import axios from '~/plugins/axios'

function rankPoints(input) {
	if (input) {
		return '\u2713'
	}
	return ''
}

function prediction(num) {
	const percent = Math.round((num - 0.5) * 100)
	if (percent > 0) {
		return `(${Math.round(num * 100)}%)`
	}
	return ''
}

export default {
	async asyncData({ params }) {
		const { data } = await axios.get('/matches/' + params.key)
		return { matches: data }
	},
	methods: {
		rankPoints: rankPoints,
		displayPrediction: prediction
	}
}
</script>
