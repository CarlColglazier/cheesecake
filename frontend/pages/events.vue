<template>
    <content>
	<h1>Events</h1>
	<ul v-for="event in events">
	    <li><nuxt-link :to="'/event/' + event.key">{{ event.short_name }}</nuxt-link></li>
	</ul>
    </content>
</template>

<script>
 export default {
     layout: 'default',
     async asyncData(context) {
         const resp = await context.app.fetch('/events');
         const js = await resp.json();
         return { events: js };
     }
 }
</script>
