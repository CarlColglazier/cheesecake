<template>
	<v-layout wrap>
		<v-flex xs12>
			<v-card>
				<v-card-title class="title">
					Welcome to Cheesecake
				</v-card-title>
				<v-card-text>
					Cheesecake is an evidence-based approach to FRC predictions and
					scouting.
				</v-card-text>
			</v-card>
		</v-flex>
		<v-flex xs12 sm6>
			<v-card class="mx-auto" tile>
				<v-card-title class="title">
					Events
				</v-card-title>
				<v-card-text>
					<v-list>
						<v-list-item v-for="event in events" :key="event.key">
							<v-list-item-title>
								<nuxt-link :to="'/event/' + event.key">
									{{ event.short_name }}
								</nuxt-link>
							</v-list-item-title>
						</v-list-item>
					</v-list>
				</v-card-text>
			</v-card>
		</v-flex>
		<v-flex xs12 sm6>
			<v-card class="mx-auto">
				<v-card-title class="title">
					Current Ratings
				</v-card-title>
				<v-card-text>
					<ol id="v-for-object">
						<li v-for="(rating, index) in ratings.slice(0, 50)" :key="index">
							{{ rating.team }} ({{ rating.score }})
						</li>
					</ol>
				</v-card-text>
			</v-card>
		</v-flex>
	</v-layout>
</template>

<script>
import axios from '~/plugins/axios'
export default {
	layout: 'default',
	async asyncData() {
		const { data } = await axios.get('/elo')
		const values = []
		for (const [key, value] of Object.entries(data)) {
			values.push({ team: key, score: value })
		}
		values.sort((a, b) => {
			return b.score - a.score
		})
		const res = await axios.get('/events')
		return { ratings: values, events: res.data }
	}
}
</script>
