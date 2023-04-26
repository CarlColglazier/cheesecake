
import {EventDataType, TeamSimsType, TeamCalculation, EVType} from '../types';

function mean_bool(e: TeamSimsType, key: string) {
  const counts = Object.keys(e).map(k => {
    return e[k][key]["true"]
  })
  const s = counts.reduce((a,b) => a + b, 0);
  return s / counts.length;
}

function logit(p: number) {
  return Math.log(p / (1 - p))
}

function from_log(p: number) {
  return Math.exp(p) / (1 + Math.exp(p))
}

function calculate_logit(team_sims: TeamSimsType, team: string, key: string, mean: number) {
  return Math.exp(logit(team_sims[team][key]["true"]/1000) - logit(mean))
}

function score_sum(team_sims: TeamSimsType, team: string, key: string) {
  return Object.entries(team_sims[team][key]).map(([k, v], i) => {
    return parseInt(k) * v;
  }).reduce((a,b) => a + b, 0);
}

function calculate_auto(team_sims: TeamSimsType, team: string) {
  const autoT = score_sum(team_sims, team, "auto_countT");
  //const autoM = score_sum(team_sims, team, "auto_countM");
  //const autoB = score_sum(team_sims, team, "auto_countB");
  const auto_charge = score_sum(team_sims, team, "auto_charge");

  return (5*autoT 
    //+ 4*autoM + 3*autoB 
    + auto_charge)/1000;
}

function calculate_tele(team_sims: TeamSimsType, team: string) {
  const T = score_sum(team_sims, team, "tele_countT");
  //const M = score_sum(team_sims, team, "tele_countM");
  //const B = score_sum(team_sims, team, "tele_countB");
  const links = score_sum(team_sims, team, "link_count")
  return (3.5*T 
    //+ 3*M + 2*B 
    + 5*links)/1000;
}

function calculate_endgame(team_sims: TeamSimsType, team: string) {
  return score_sum(team_sims, team, "endgame")/1000;
}

function bg_classname_logit(n: number) {
  const op = Math.round(10*Math.round(10*Math.min(Math.abs(n-1), 1)));
  if (n > 1) {
    return `bg-blue-50 bg-opacity-${op}`
  } else {
    return `bg-red-50 bg-opacity-${op}`
  }
}

function bg_classname(n: number, mm: number) {
  const op = Math.round(10*Math.round(10*Math.abs(n)/mm));
  if (n > 1) {
    return `bg-blue-50 bg-opacity-${op}`
  } else {
    return `bg-red-50 bg-opacity-${op}`
  }
}

function mean_ev(ev: EVType) {
  const count = ev["bcount"].map((v, i) => {
    return v * ev["points"][i];
  }).reduce((a,b) => a + b, 0);
  return count / 1_000;
}

// https://stackoverflow.com/questions/45309447/calculating-median-javascript
const median = (arr: number[]): number => {
  const s = [...arr].sort((a, b) => a - b);
  const mid = Math.floor(s.length / 2);
  return s.length % 2 === 0 ? ((s[mid - 1] + s[mid]) / 2) : s[mid];
};

const absmax = (arr: number[], med: number): number => {
  return Math.max(...arr.map(n => Math.abs(n - med)));
}

const EVTable: React.FC<EventDataType> = ({ ev, team_sims }) => {
  const mean_activations = mean_bool(team_sims, "activation");
  const mean_sustainability = mean_bool(team_sims, "sustainability");

  const team_auto_scores = ev.map(e => {
    return calculate_auto(team_sims, e["team"])
  });
  const median_auto = median(team_auto_scores);
  const minmax_auto = absmax(team_auto_scores, median_auto);

  const team_tele_scores = ev.map(e => {
    return calculate_tele(team_sims, e["team"])
  });
  const median_tele = median(team_tele_scores);
  const minmax_tele = absmax(team_tele_scores, median_tele);

  const team_endgame_scores = ev.map(e => {
    return calculate_endgame(team_sims, e["team"])
  });
  const median_endgame = median(team_endgame_scores);
  const minmax_endgame = absmax(team_endgame_scores, median_endgame);


  //let evs: {[key:string]: number} = {}
  let tabledata: TeamCalculation = {};
  for (const o of ev) {
    const team = o["team"];
    tabledata[team] = {
      "ev_mean": mean_ev(o),
      "auto": calculate_auto(team_sims, team),
      "tele": calculate_tele(team_sims, team),
      "endgame": calculate_endgame(team_sims, team),
      "activation": team_sims[team]["activation"]["true"],
      "sustainability": team_sims[team]["sustainability"]["true"]
    }
  }

  return (
    <table className='table-auto w-full'>
      <thead>
        <tr className='text-right'>
          <th>Team</th>
          <th>MPAR</th>
          {/*<th>Act. RP</th>
          <th>Sus. RP</th>*/}
          <th>Auto</th>
          <th>Tele</th>
          {/*<th>End</th>*/}
        </tr>
      </thead>
      <tbody className='text-right'>
        {Object.entries(tabledata).sort((a, b) => { return b[1]["ev_mean"] - a[1]["ev_mean"]})
        .map(([team, e], i) => <tr key={i}>
          <td>{team}</td>
          <td>{e["ev_mean"].toFixed(1)}</td>
          {/*
          <td className={bg_classname_logit(calculate_logit(team_sims, team, "activation", mean_activations))}>
            {Math.round(100*e["activation"])}%
          </td>
          <td className={bg_classname_logit(calculate_logit(team_sims, team, "sustainability", mean_sustainability))}>
          {Math.round(100*e["sustainability"])}%
          </td>
          */}
          <td className={bg_classname(e["auto"]-median_auto, minmax_auto)}>
            {Math.round(e["auto"])}
          </td>
          <td className={bg_classname(e["tele"]-median_tele, minmax_tele)}>
            {Math.round(e["tele"])}
          </td>
          {/*<td className={bg_classname(e["endgame"]-median_endgame, minmax_endgame)}>
            {Math.round(e["endgame"])}
            //Math.round(score_sum(team_sims, team, "endgame")/1000)
          </td>*/}
          </tr>)}
      </tbody>
    </table>
  )
}

export default EVTable;

