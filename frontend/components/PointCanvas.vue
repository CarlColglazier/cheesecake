<template>
	<canvas width="100" height="50">
	</canvas>
</template>

<script>
function draw() {
	var c = document.getElementById(this.team+'-point');
  var ctx = c.getContext("2d");
	ctx.fillStyle = "#999";
	/*
	for (var d of this.std) {
	}
	*/
	ctx.fillStyle = "#000";
	for (var d of this.mean) {
		//ctx.fillRect(d.match, 50 - d.forecast * 50, 5, 10);
		ctx.fillRect(d.match, 50 - d.forecast, 1, d.forecast);
	}
}

export default {
	props: ['team', 'std', 'mean'],
	methods: {
		draw: draw
	},
	mounted() {
		this.draw()
	}
}
</script>
