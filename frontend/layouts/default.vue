<template>
	<div>
		<b-navbar type="light" variant="faded">
			<b-navbar-brand href="/">Cheesecake Live</b-navbar-brand>
			<b-collapse id="nav-collapse" is-nav>
				<b-navbar-nav>
					<b-nav-item href="/events">Events</b-nav-item>
				</b-navbar-nav>
			</b-collapse>
		</b-navbar>
		<b-container>
			<nuxt/>
		</b-container>
	</div>
</template>
