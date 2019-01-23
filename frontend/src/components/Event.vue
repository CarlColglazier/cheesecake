<template>
  <div>
    <Header/>
    <h2>{{ this.$route.params.key }}</h2>
    <table>
      <thead>
        <tr><th>Team</th><th>Win Rate</th></tr>
      </thead>
      <tbody>
        <tr v-for="(sim, index) in simulation" :key="index" v-if="index < 8">
          <td>{{sim.key}}</td>
          <td class="right">{{sim.mean | round(2)}}</td>
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
    this.$http.get(`simulate/${this.$route.params.key}`)
      .then(data => {
        return data.json()
      }).then(data => {
        this.simulation = data
      })
  }
}
</script>

<style>
.right {
  text-align: right;
}
</style>
