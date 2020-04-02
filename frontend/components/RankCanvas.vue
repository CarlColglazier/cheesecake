<template>
	<canvas width="100" height="50">
	</canvas>
</template>

<script>
function draw() {
	var c = document.getElementById(this.team+'-rank');
  var ctx = c.getContext("2d");
	ctx.fillStyle = "#999";
	for (var d of this.cap) {
		ctx.fillRect(
			d.match,
			Math.round(50 - d.forecast * 50),
			1,
			Math.round(d.forecast * 50)
		);
	}
	ctx.fillStyle = "#000";
	for (var d of this.leader) {
		ctx.fillRect(
			d.match,
			Math.round(50 - d.forecast * 50),
			1,
			Math.round(d.forecast * 50)
		);
	}
}

export default {
	props: ['team', 'cap', 'leader'],
	methods: {
		draw: draw
	},
	mounted() {
		this.draw()
	}
}
</script>
