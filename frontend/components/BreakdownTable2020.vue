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
						<nuxt-link :to="'/team/' + team + '/2020'">
							{{ team.substring(3) }}
						</nuxt-link>
					</span>
				</b-td>
				<b-td>{{ score(d.alliances.red.alliance.score) }}</b-td>
				<b-td>({{ roundPred(d.predictions['eloscore2020'].prediction.red) }})</b-td>
				<b-td>{{ rankPoints(rP(d, 'red', 'shieldEnergizedRankingPoint')) }}</b-td>
				<b-td>{{ rpPrediction(d, "red", "shieldeng") }}</b-td>
				<b-td>{{ rankPoints(rP(d, 'red', 'shieldOperationalRankingPoint')) }}</b-td>
				<b-td>{{ rpPrediction(d, "red", "shieldop") }}</b-td>
			</b-tr>
			<b-tr>
				<b-td></b-td>
				<b-td>
					<span
						v-for="team in d.alliances.blue.teams"
						:key="team"
					>
						<nuxt-link :to="'/team/' + team + '/2020'">
							{{ team.substring(3) }}
						</nuxt-link>
					</span>
				</b-td>
				<b-td>{{ score(d.alliances.blue.alliance.score) }}</b-td>
				<b-td>({{ roundPred(d.predictions['eloscore2020'].prediction.blue) }})</b-td>
				<b-td>{{ rankPoints(rP(d, 'blue', 'shieldEnergizedRankingPoint')) }}</b-td>
				<b-td>{{ rpPrediction(d, "blue", "shieldeng") }}</b-td>
				<b-td>{{ rankPoints(rP(d, 'blue', 'shieldOperationalRankingPoint')) }}</b-td>
				<b-td>{{ rpPrediction(d, "blue", "shieldop") }}</b-td>
			</b-tr>
		</b-tbody>
	</b-table-simple>
</template>

<script>
function rP(d, color, prop) {
	if (!('score_breakdown' in d.match)) {
		return '';	
	}
	if (d.match.score_breakdown === null) {
		return '';
	}
	return d.match.score_breakdown[color][prop];
}

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

function score(num) {
	if (num === -1) {
		return '';
	}
	return `${num}`
}

function rpPrediction(match, color, key) {
	if (match.match.comp_level != "qm") {
		return '';
	}
	return prediction(match.predictions[key].prediction[color]);
}

export default {
	methods: {
		rankPoints: rankPoints,
		displayPrediction: prediction,
		roundPred: roundPred,
		rP: rP,
		score: score,
		rpPrediction: rpPrediction
	},
	props: ['matches']
}
</script>
