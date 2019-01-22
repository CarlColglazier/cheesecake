<template>
  <div>
    <Header/>
    <h2>{{ this.$route.params.key }}</h2>
    <table>
      <thead>
        <tr><th>Team</th><th>Wins</th></tr>
      </thead>
      <tbody>
        <tr v-for="(sim, index) in simulation" :key="index">
          <td>{{index}}</td>
          <td>{{sim.mean}}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import Header from '@/components/Header'

export default {
  name: 'Event',
  components: { Header },
  data () {
    return {
      simulation: []
    }
  },
  mounted () {
    fetch(`http://localhost:5000/api/simulate/${this.$route.params.key}`)
      .then(data => {
        return data.json()
      }).then(data => {
        this.simulation = data
      })
  }
}
</script>
