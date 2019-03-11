<template>
<tbody>
  <tr>
    <td>{{name}}</td>
    <td v-for="team in red_teams" :key="team">
      {{team}}
    </td>
    <td v-if="match.alliances.red.score != -1"
        class="red right">
        {{match.alliances.red.score}}
    </td>
    <td v-else class="red center">-</td>
    <td class="right" :class="{
      strike: (match.predictions.EloScorePredictor < 0.5 && match.winning_alliance =='red') || (match.predictions.EloScorePredictor > 0.5 && match.winning_alliance =='blue')
    }">
      {{match.predictions.EloScorePredictor | rounds(2) }}
    </td>
    <td class="right">
      <span v-if="match.score_breakdown.red.habDockingRankingPoint">✔</span>
      {{match.predictions.habDockingRankingPointred | rounds(2)}}
    </td>
    <td class="right">
      <span v-if="match.score_breakdown.red.completeRocketRankingPoint">✔</span>
      {{match.predictions.completeRocketRankingPointred | rounds(2)}}
    </td>
  </tr>
  <tr>
    <td></td>
    <td v-for="team in blue_teams" :key="team">
      {{team}}
    </td>
    <td v-if="match.alliances.blue.score != -1"
        class="blue right">
        {{match.alliances.blue.score}}
    </td>
    <td v-else class="blue center">-</td>
    <td class="right" :class="{
                strike: (match.predictions.EloScorePredictor < 0.5 && match.winning_alliance =='red') || (match.predictions.EloScorePredictor > 0.5 && match.winning_alliance =='blue')
    }">
      {{1 - match.predictions.EloScorePredictor | rounds(2) }}
    </td>
    <td class="right">
      <span v-if="match.score_breakdown.blue.habDockingRankingPoint">✔</span>
      {{match.predictions.habDockingRankingPointblue | rounds(2)}}
    </td>
    <td class="right">
      <span v-if="match.score_breakdown.blue.completeRocketRankingPoint">✔</span>
      {{match.predictions.completeRocketRankingPointblue | rounds(2)}}
    </td>
  </tr>
</tbody>
</template>

<script>
export default {
  name: 'PredictionRow',
  props: ['match'],
  computed: {
    name () {
      return this.match.key.split('_')[1]
    },
    red_teams () {
      return this.match.alliances.red.team_keys.map(key => {
        return key.replace('frc', '')
      })
    },
    blue_teams () {
      return this.match.alliances.blue.team_keys.map(key => {
        return key.replace('frc', '')
      })
    }
  }
}
</script>
