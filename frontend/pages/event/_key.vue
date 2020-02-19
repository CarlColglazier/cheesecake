<template>
	<content>
		<h1>{{ $nuxt.$route.params.key }}</h1>
		<ForecastTable v-bind:forecasts="forecasts" />
		<section v-if="matches && matches.length > 0">
			<p v-if="yearMatch(2020)">Note: this model has still not been fully calibrated for the 2020 game.</p>
			<h2>Breakdown for {{ $nuxt.$route.params.key }}</h2>
			<BreakdownTable2019 v-if="yearMatch(2019)" v-bind:matches="matches" />
			<BreakdownTable2020 v-if="yearMatch(2020)" v-bind:matches="matches" />
		</section>
		<secion v-else>
			<p>No schedule for this event quite yet.</p>
		</secion>
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
import ForecastTable from "~/components/ForecastTable.vue";

function yearMatch(year) {
	return this.$nuxt.$route.params.key.substring(0, 4) == year;
}

function topTeams(forecast) {
	var max_match = Math.max.apply(Math, forecast.map(function(f) { return f.match }));
	var max_fore = forecast.filter(function(f) { return f.match === max_match });
	return max_fore.map(function(f) { return f.team });
}

export default {
	layout: 'default',
	async asyncData(context) {
		try {
			const dataf = await context.app.fetch(`/matches/${context.params.key}`)
			const data = await dataf.json()
			const foref = await context.app.fetch(`/forecasts/${context.params.key}`)
			const fore = await foref.json()
			return { matches: data, forecasts: fore }
		} catch (e) {
			console.error(e)
			return { matches: [], forecasts: [] }
		}
	},
	mounted() {
		//this.renderChart()
	},
	components: {
		BreakdownTable2019: BreakdownTable2019,
		BreakdownTable2020: BreakdownTable2020,
		ForecastTable: ForecastTable
	},
	methods: {
		yearMatch: yearMatch
	}
}
</script>
