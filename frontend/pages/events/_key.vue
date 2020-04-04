<template>
  <content>
		<h1>Events</h1>
		<b-table-simple>
			<b-thead>
				<b-tr>
					<b-th>Event</b-th>
					<b-th>End Date</b-th>
				</b-tr>
			</b-thead>
			<b-tbody>
				<b-tr v-for="event in orderedEvents">
					<b-td>
						<nuxt-link :to="'/event/' + event.key">{{ event.short_name }}</nuxt-link>
					</b-td>
					<b-td>{{ event.end_date }}</b-td>
				</b-tr>
			</b-tbody>
		</b-table-simple>
  </content>
</template>

<script>
import _ from 'lodash';

export default {

  layout: 'default',
  async asyncData(context) {
    const resp = await context.app.fetch(`/events/${context.params.key}`);
    const js = await resp.json();
    return { events: js };
  },
	computed: {
		orderedEvents: function () {
			return _.orderBy(this.events, 'end_date')
		}
	}
}
</script>
