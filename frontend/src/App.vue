<template>
  <div id="app" class="container">
    <router-view/>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
 h1, h2, h3, h4, h5, h6, th {
     font-family: 'ABeeZee', sans-serif;
 }
 body {
     font-family: 'Lato', sans-serif;
 }
 .container {
     max-width: 64em;
     margin: 0 auto;
     padding: 1em;
 }
 a, a:visited {
     color: #44A;
 }
</style>
