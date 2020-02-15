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
					RP1
				</b-th>
				<b-th colspan="2">
					RP2
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
