<template>
  <b-row>
		<b-col>
			<b-jumbotron>
				<template v-slot:header>Cheesecake</template>
				<template v-slot:lead>
					Cheesecake is a live forecasting system for the <em>FIRST</em> Robotics Competition.
				</template>
				<hr class="my-4">
				<p>Cheesecake uses a number of models to predict future results of FRC matches. It records these predictions as probabilities, which can be analyzed for accuracy and calibration.</p>
			</b-jumbotron>
			<section>
				<h3>Upcoming Events</h3>
				<ul>
					<li v-for="event in filterEvents(events)">
						<nuxt-link :to="'/event/' + event.key">{{ event.short_name }}</nuxt-link>
					</li>
				</ul>
			</section>
			<section>
				<h3>Credits</h3>
				<p>The underlying code for this website comes from Carl Colglazier <a href="https://twitter.com/carlcolglazier">(twitter)</a>.</p>
				<p>Parts of the Elo model share similarities with the work of Caleb Sykes <a href="https://github.com/inkling16/">(github)</a>.</p>
				<p>The entire project is <a href="https://github.com/carlcolglazier/cheesecake">open source</a>.</p>
			</section>
			<section>
				<h3>Fun Numbers for Nerds</h3>
				<p>Brier: {{ brier.score }}. Predicted {{ brier.correct }} / {{ brier.count }} ({{ Math.round(100 * brier.correct / brier.count) }}%) matches.</p>
			</section>
		</b-col>
  </b-row>
</template>

<script>
function eventUpcoming(event) {
	let date = Date.parse(event.end_date)
	let now = Date.now()
	return date - now > 0 &&
				 date - now < 1000 * 60 * 60 * 24 * 7;
}

function filterEvents(events) {
	return events.filter(m => {
		return eventUpcoming(m);
	});
}

export default {
	layout: 'default',
	async asyncData(context) {
		const ev = await context.app.fetch("/events/2020");
		const ev_j = await ev.clone().json();
		const b_rest = await context.app.fetch("/brier");
		const b_rest_j = await b_rest.clone().json();
		return { events: ev_j, brier: b_rest_j };
	},
	methods: {
		filterEvents: filterEvents
	}
}
</script>
