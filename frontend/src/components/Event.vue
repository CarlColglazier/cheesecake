<template>
  <div>
    <Header/>
    <section v-if="matches == null">
      <Error message="Event not found" />
    </section>
    <section v-else>
      <h2>{{ this.$route.params.key }}</h2>
      <a :href="tbahref">
        The Blue Alliance
      </a>
      <h3>Matches</h3>
      <table v-if="matches.length > 0">
        <thead>
          <tr>
            <th>Match</th>
            <th colspan="3">Teams</th>
            <th>Score</th>
            <th>Win</th>
            <th>Dock</th>
            <th>Rocket</th>
          </tr>
        </thead>
        <PredictionRow v-for="match in matches" :match="match" :key="match.key" />
      </table>
      <p v-else>
        Matches have not been released yet for this event.
        Please check back later.
      </p>
    </section>
    <section v-if="matches.length > 0">
      <h3>Rankings (Predicted)</h3>
      <table>
        <thead><tr><th>Rank</th><th>Team</th><th>Est. Rank Points</th></tr></thead>
        <tbody>
          <tr v-for="(key, index) in orderedrank" :key="index">
            <td>{{index + 1}}</td><td>{{key[0]}}</td><td class="right">{{key[1] | round(1) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<script>
import Header from '@/components/Header'
import Error from '@/components/Error'
import PredictionRow from '@/components/PredictionRow'

export default {
  name: 'Event',
  components: { Header, Error, PredictionRow },
  data () {
    return {
      matches: [],
      rankings: {}
    }
  },
  computed: {
    tbahref () {
      return `https://www.thebluealliance.com/event/${this.$route.params.key}`
    },
    orderedrank () {
      // https://stackoverflow.com/questions/25500316/sort-a-dictionary-by-value-in-javascript
      var dict = this.rankings
      var items = Object.keys(dict).map(function (key) {
        return [key, dict[key]]
      })
      items.sort(function (first, second) {
        return second[1] - first[1]
      })
      return items
    }
  },
  mounted () {
    this.$http.get(`matches/${this.$route.params.key}`)
      .then(data => {
        return data.json()
      }).then(data => {
        this.matches = data
      }).catch(_ => {
        this.matches = null
      })
    this.$http.get(`rankings/${this.$route.params.key}`)
      .then(data => {
        return data.json()
      }).then(data => {
        this.rankings = data
      })
  }
}
</script>

<style>
 .right {
     text-align: right;
 }
 .center {
     text-align: center;
 }
 .red {
     background-color: #FEE;
 }
 .blue {
     background-color: #EEF;
 }
 th, td {
     min-width: 2em;
     padding: .5em;
 }

 progress {
     background: #CCF;
 }

 progress[value] {
     /* Reset the default appearance */
     -webkit-appearance: none;
     -moz-appearance: none;
     appearance: none;
     border: none;
 }

 progress::-moz-progress-bar {
     color: red;
 }
 progress::after { content: attr(value); }

 .strike {
     text-decoration: line-through;
 }

 section {
     display: inline-block;
     padding-right: 7.5em;
  }
</style>
