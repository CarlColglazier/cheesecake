<template>
	<v-app>
		<!--<v-navigation-drawer v-model="drawer" app>
		</v-navigation-drawer>-->
		<v-app-bar>
			<!--<v-toolbar-side-icon @click="drawer = !drawer" />-->
			<nuxt-link to="/">
				<v-toolbar-title>Cheesecake</v-toolbar-title>
			</nuxt-link>
			<v-spacer />
		</v-app-bar>
		<!-- Sizes your content based upon application components -->
		<v-content>
			<!-- Provides the application the proper gutter -->
			<v-container grid-list-md text-center>
				<router-view />
			</v-container>
		</v-content>
		<v-footer app>
			<!-- -->
		</v-footer>
	</v-app>
</template>

<script>
export default {
	data() {
		return {
			clipped: false,
			drawer: false,
			fixed: false,
			items: [
				{ icon: 'apps', title: 'Welcome', to: '/' },
				{ icon: 'bubble_chart', title: 'Inspire', to: '/inspire' }
			],
			miniVariant: false,
			right: true,
			rightDrawer: false,
			title: 'Vuetify.js'
		}
	}
}
</script>
