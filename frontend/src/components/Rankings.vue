<template>
  <div>
    <Header/>
    <div>
      <h2>Current Top 25</h2>
      <ol>
        <li v-for="team in rankings" :key="team[0]">
          {{team[0]}}
        </li>
      </ol>
    </div>
  </div>
</template>

<script>
import Header from '@/components/Header'

export default {
  name: 'Rankings',
  components: { Header },
  data () {
    return {
      rankings: []
    }
  },
  mounted () {
    this.$http.get(`teams/rankings`)
      .then(data => {
        return data.json()
      }).then(data => {
        this.rankings = data
      })
  }
}
</script>
