<template>
	<h1>Homepage</h1>
</template>

<script>
export default {
	layout: 'default',
	async asyncData() {
	      const resp = await fetch('http://backend:8080/events');
	      const js = await resp.json();
	      console.log(JSON.stringify(js));
	}
}
</script>
