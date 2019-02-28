<template>
  <div>
    <Header/>
    <h2>Upcoming Events</h2>
    <ul class="menulist">
      <li v-for="event in events" :key="event.key">
        <router-link :to="`/event/${event.key}`">
          <a>{{ event.name }}</a>
        </router-link>
      </li>
    </ul>
  </div>
</template>

<script>
import Header from '@/components/Header'

export default {
  name: 'Events',
  components: { Header },
  data () {
    return {
      events: []
    }
  },
  mounted () {
    this.$http.get('events/upcoming')
      .then(data => {
        return data.json()
      }).then(data => {
        this.events = data
      })
  }
}
</script>

<style>
 .menulist {
     padding: 0;
 }
 .menulist li {
     list-style-type: none;
     margin: 1em .1em;
 }
</style>
