<template>
	<content>
		<h1>{{ $nuxt.$route.params.key }}</h1>
		<p>The race for first place</p>
		<div id="chart"></div>
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

import * as d3 from 'd3';

function yearMatch(year) {
	return this.$nuxt.$route.params.key.substring(0, 4) == year;
}

function topTeams(forecast) {
	var max_match = Math.max.apply(Math, forecast.map(function(f) { return f.match }));
	var max_fore = forecast.filter(function(f) { return f.match === max_match });
	return max_fore.map(function(f) { return f.team });
}

function renderChart() {
	var dat = this.forecasts;
	var top_teams = topTeams(dat);

	// set the dimensions and margins of the graph
	var margin = {top: 10, right: 0, bottom: 30, left: 20},
			width = 150 - margin.left - margin.right,
			height = 100 - margin.top - margin.bottom;

	var sumstat = d3.nest()
									.key(function(d) { return d.team; })
									.entries(dat);

	// append the svg object to the body of the page
	var svg = d3.select("#chart")
							.selectAll("uniqueChart")
							.data(sumstat)
							.enter()
							.append("svg")
							.attr("width", width + margin.left + margin.right)
							.attr("height", height + margin.top + margin.bottom)
							.append("g")
							.attr("transform",
										"translate(" + margin.left + "," + margin.top + ")");

	// Add X axis --> it is a date format
  var x = d3.scaleLinear()
						.domain(d3.extent(dat, function(d) { return d.match; }))
						.range([ 0, width ]);
	
  svg.append("g")
     .attr("transform", "translate(0," + height + ")")
     //.call(d3.axisBottom(x));

  // Add Y axis
  var y = d3.scaleLinear()
	//.domain([0, d3.max(dat, function(d) { return +d.forecast; })])
	          .domain([0, 1])
						.range([ height, 0 ]);
  svg.append("g")
     //.call(d3.axisLeft(y));

	svg.append("path")
	//.datum(dat)
		 .attr("fill", "black")
		 .attr("stroke", "black")
		 .attr("stroke-width", 1.5)
		 .attr("d", function(d) {
			 return d3.area()
								.x(function(d) { return x(d.match) })
								.y0(y(0))
								.y1(function(d) { return y(d.forecast) })
			 (d.values)
		 })

	svg
     .append("text")
     .attr("text-anchor", "start")
     .attr("y", 5)
     .attr("x", 0)
     .text(function(d){ return(d.key)})
		 .style("fill", "black")
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
		this.renderChart()
	},
	components: {
		BreakdownTable2019: BreakdownTable2019,
		BreakdownTable2020: BreakdownTable2020
	},
	methods: {
		yearMatch: yearMatch,
		renderChart: renderChart
	}
}
</script>
