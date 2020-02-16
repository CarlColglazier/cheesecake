<template>
	<b-table-simple>
		<b-thead>
			<b-tr>
				<b-th>Key</b-th>
				<b-th colspan="1">
					Teams
				</b-th>
				<b-th colspan="2">
					Score
				</b-th>
				<b-th colspan="2">
					Energized
				</b-th>
				<b-th colspan="2">
					Shield
				</b-th>
				<b-th>
					Rank Points
				</b-th>
			</b-tr>
		</b-thead>
		<b-tbody v-for="d in matches" :key="d.match.key">
			<b-tr>
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
					>
						{{ team.substring(3) }}
					</span>
				</b-td>
				<b-td>{{ d.alliances.red.alliance.score }}</b-td>
				<b-td>({{ roundPred(d.predictions.elo_score.prediction.red) }})</b-td>
				<b-td>{{ rankPoints(d.match.score_breakdown.red.shieldEnergizedRankingPoint) }}</b-td>
				<b-td>({{ roundPred(d.predictions.energized.prediction.red) }})</b-td>
				<b-td>{{ rankPoints(d.match.score_breakdown.red.shieldOperationalRankingPoint) }}</b-td>
				<b-td>({{ roundPred(d.predictions.shield.prediction.red) }})</b-td>
			</b-tr>
			<b-tr>
				<b-td></b-td>
				<b-td>
					<span
						v-for="team in d.alliances.blue.teams"
									 :key="team"
					>
						{{ team.substring(3) }}
					</span>
				</b-td>
				<b-td>{{ d.alliances.blue.alliance.score }}</b-td>
				<b-td>({{ roundPred(d.predictions.elo_score.prediction.blue) }})</b-td>
				<b-td>{{ rankPoints(d.match.score_breakdown.blue.shieldEnergizedRankingPoint) }}</b-td>
				<b-td>({{ roundPred(d.predictions.energized.prediction.blue) }})</b-td>
				<b-td>{{ rankPoints(d.match.score_breakdown.blue.shieldOperationalRankingPoint) }}</b-td>
				<b-td>({{ roundPred(d.predictions.shield.prediction.blue) }})</b-td>
			</b-tr>
		</b-tbody>
	</b-table-simple>
</template>

<script>
function rankPoints(input) {
	if (input) {
		return '\u2713'
	}
	return ''
}

function prediction(num) {
	return `(${Math.round(num * 100)}%)`
}

function roundPred(num) {
	return `${Math.round(num * 100)}%`
}

export default {
	methods: {
		rankPoints: rankPoints,
		displayPrediction: prediction,
		roundPred: roundPred
	},
	props: ['matches']
}
</script>
