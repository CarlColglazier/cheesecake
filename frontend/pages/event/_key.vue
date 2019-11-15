<template>
	<div>
		<p>{{ this.$route.params.key }}, Welcome to Cheesecake 2</p>
		<v-simple-table>
			<thead>
				<tr>
					<th>Key</th>
					<th colspan="3">Red</th>
					<th colspan="3">Blue</th>
					<th colspan="2">Score</th>
				</tr>
			</thead>
			<tbody>
				<tr v-for="d in matches" :key="d.Match.Key">
					<td>{{ d.Match.Key }}</td>
					<td v-for="team in d.Alliances.red.Teams" :key="team">
						{{ team.substring(3) }}
					</td>
					<td v-for="team in d.Alliances.blue.Teams" :key="team">
						{{ team.substring(3) }}
					</td>
					<td>{{ d.Alliances.red.Alliance.Score }}</td>
					<td>{{ d.Alliances.blue.Alliance.Score }}</td>
				</tr>
			</tbody>
		</v-simple-table>
	</div>
</template>

<script>
import axios from '~/plugins/axios'
export default {
	async asyncData({ params }) {
		const { data } = await axios.get('/matches/' + params.key)
		return { matches: data }
	}
}
</script>
