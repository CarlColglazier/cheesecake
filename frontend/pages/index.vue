<template>
	<v-layout wrap>
		<v-flex xs12>
			<v-card>
				<v-card-title class="title">
					Welcome to Cheesecake
				</v-card-title>
			</v-card>
		</v-flex>
		<v-flex xs12 sm6>
			<v-card>
				<v-card-title class="title">
					Events
				</v-card-title>
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
							{{ rating.team }}
						</li>
					</ol>
				</v-card-text>
			</v-card>
		</v-flex>
	</v-layout>
</template>

<script>
import axios from 'axios'
export default {
	layout: 'default',
	async asyncData() {
		const { data } = await axios.get('http://localhost:8080/elo')
		const values = []
		for (const [key, value] of Object.entries(data)) {
			values.push({ team: key, score: value })
		}
		values.sort((a, b) => {
			return b.score - a.score
		})
		return { ratings: values }
	}
}
</script>
