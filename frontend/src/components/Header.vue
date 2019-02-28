<template>
  <header>
    <h1>Cheesecake</h1>
    <ul>
      <router-link tag="li" to="/"><a>Home</a></router-link>
      <router-link tag="li" to="/events"><a>Events</a></router-link>
    </ul>
  </header>
</template>

<style>
 header {
     text-align: center;
 }
 header ul {
     padding: 0;
 }
 header ul li {
     display: inline-block;
     margin: 0 auto;
 }
 header ul li a {
     padding: 1em;
 }
</style>
