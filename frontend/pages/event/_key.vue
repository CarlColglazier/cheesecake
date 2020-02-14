<template>
	<content>
		<h1>{{ $nuxt.$route.params.key }}</h1>
		<h2>Breakdown for {{ $nuxt.$route.params.key }}</h2>
		<BreakdownTable2019 v-if="yearMatch(2019)" v-bind:matches="matches" />
		<BreakdownTable2020 v-if="yearMatch(2020)" v-bind:matches="matches" />
	</content>
</template>

<style>
 span.break {
		 display: block;
 }
</style>

<script>
import BreakdownTable2019 from "~/components/BreakdownTable2019.vue";
import BreakdownTable2020 from "~/components/BreakdownTable2020.vue";

function yearMatch(year) {
	return this.$nuxt.$route.params.key.substring(0, 4) == year;
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
	components: {
		BreakdownTable2019: BreakdownTable2019,
		BreakdownTable2020: BreakdownTable2020
	},
	methods: {
		yearMatch: yearMatch
	}
}
</script>
