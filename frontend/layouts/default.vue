<template>
	<div>
		<b-navbar type="light" variant="faded">
			<b-navbar-brand href="/"><strong>Cheesecake Live</strong> (Alpha)</b-navbar-brand>
			<b-collapse id="nav-collapse" is-nav>
				<b-navbar-nav>
					<b-nav-item-dropdown text="Events">
						<b-nav-item href="/events/2020">2020</b-nav-item>
						<b-nav-item href="/events/2019">2019</b-nav-item>
					</b-nav-item-dropdown>
				</b-navbar-nav>
			</b-collapse>
		</b-navbar>
		<b-container class="mt-5">
			<nuxt/>
		</b-container>
	</div>
</template>
