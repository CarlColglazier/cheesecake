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
					Hab
				</b-th>
				<b-th colspan="2">
					Rocket
				</b-th>
				<b-th>
					Rank Points
				</b-th>
			</b-tr>
		</b-thead>
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
