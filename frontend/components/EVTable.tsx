
import {EVType, TeamSimType1, TeamSimsType, EventDataType, TeamSummaryType} from '../types';
import quartile from '../util';

function med(o: TeamSimType1) {
    const keys = Object.keys(o);
    const vals = Object.values(o);
    return quartile(keys, vals, 0.5);
  }
  
  function team_summery(e: EVType) {
    return { "team":e.team, "median":quartile(e.points, e.bcount, 0.5), "qlow":quartile(e.points, e.bcount, 0.1), "qhigh":quartile(e.points, e.bcount, 0.9) }
  }
  
  const EVTable: React.FC<EventDataType> = ({ ev, team_sims }) => {
    const neworder = ev.map(team_summery).sort(function(a: TeamSummaryType, b: TeamSummaryType) { return b.median - a.median });
    const neworderteams = neworder.map(function(e: {team: string}) { return e.team });
    return (
      <table className='table-auto w-full'>
        <thead>
          <tr className='text-right'>
            <th>Team</th>
            <th>Median</th>
            <th>Low</th>
            <th>High</th>
            <th>Taxi</th>
            <th>Traversal</th>
            <th>U(A)</th>
            <th>L(A)</th>
            <th>U</th>
            <th>L</th>
          </tr>
        </thead>
        <tbody className='text-right'>
          {neworder.map((e, i) => <tr key={i}>
            <td>{e.team}</td>
            <td>{e.median}</td>
            <td>{e.qlow}</td>
            <td>{e.qhigh}</td>
            <td>{Math.round(team_sims[e.team]["taxi"]["true"]/10)}%</td>
            <td>{Math.round(team_sims[e.team]["climb"][15]/10)}%</td>
            <td>{med(team_sims[e.team]["cargoautoupper"])}</td>
            <td>{med(team_sims[e.team]["cargoautolower"])}</td>
            <td>{med(team_sims[e.team]["cargoupper"])}</td>
            <td>{med(team_sims[e.team]["cargolower"])}</td>
            </tr>)}
        </tbody>
      </table>
    )
  }
  
  export default EVTable;