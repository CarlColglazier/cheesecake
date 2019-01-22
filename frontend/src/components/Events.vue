<template>
  <div>
    <Header/>
    <li v-for="event in events" :key="event.key">
      <router-link :to="`/event/${event.key}`">
        <a>{{ event.name }}</a>
      </router-link>
    </li>
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
    this.$http.get('events/2019')
      .then(data => {
        return data.json()
      }).then(data => {
        this.events = data
      })
  }
}
</script>
