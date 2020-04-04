<template>
	<content>
		<h1>{{ $nuxt.$route.params.key }}</h1>
		<h2>Breakdown</h2>
		<BreakdownTable2020 v-bind:matches="matches" />
		<h2>Scoring</h2>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Key</b-th>
					<b-th>Auto</b-th>
					<b-th>Tele Cell</b-th>
					<b-th>Endgame</b-th>
				</b-tr>
			</b-thead>
			<b-tbody v-for="d in matches" :key="d.match.key">
				<b-tr>
					<b-td>{{ d.match.key }}</b-td>
					<b-td>{{ d.match.score_breakdown[get_color(d)]['autoPoints'] }}</b-td>
					<b-td>{{ d.match.score_breakdown[get_color(d)]['teleopCellPoints'] }}</b-td>
					<b-td>{{ d.match.score_breakdown[get_color(d)]['endgamePoints'] }}</b-td>
				</b-tr>
			</b-tbody>
		</b-table-simple>
		
	</content>
</template>

<script>
import BreakdownTable2020 from "~/components/BreakdownTable2020.vue";

function get_color(d) {
	return Object.keys(d.alliances)[0];
}

export default {
	layout: 'default',
	async asyncData(context) {
		try {
			const dataf = await context.app.fetch(`/team/${context.params.key}/2020`)
			const data = await dataf.json()
			console.log(data);
			return { matches: data }
		} catch (e) {
			console.error(e)
			return { matches: [] }
		}
	},
	components: {
		BreakdownTable2020: BreakdownTable2020
	},
	methods: {
		get_color: get_color
	}
}
</script>
