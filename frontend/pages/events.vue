<template>
    <content>
	<h1>Events</h1>
	<ul v-for="event in events">
	    <li><nuxt-link :to="'/event/' + event.key">{{ event.short_name }}</nuxt-link></li>
	</ul>
    </content>
</template>

<script>
 let url = (process.server) ? 'http://backend:8080' : 'http://localhost:8080';
 export default {
     layout: 'default',
     async asyncData() {
         const resp = await fetch(`${url}/events`);
         const js = await resp.json();
         return { events: js };
     }
 }
</script>
