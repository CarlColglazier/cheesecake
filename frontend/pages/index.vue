<template>
	<div>
		<h1>Homepage</h1>
		<p>Cheesecake is a live forecasting system for the <em>FIRST</em> Robotics Competition.</p>
		<ul v-for="event in events.slice(0, 10)">
		  <li>{{ event.short_name }}</li>
		</ul>
	</div>
</template>

<script>
export default {
	layout: 'default',
	async asyncData() {
	      const resp = await fetch('http://backend:8080/events');
	      const js = await resp.json();
	      return { events: js };
	}
}
</script>
