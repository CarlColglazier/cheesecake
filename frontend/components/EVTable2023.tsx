
import {EventDataType} from '../types';
 
const EVTable: React.FC<EventDataType> = ({ ev, team_sims }) => {
  return (
    <table className='table-auto w-full'>
      <thead>
        <tr className='text-right'>
          <th>Team</th>
          <th>Activation</th>
          <th>Sustainability</th>
          <th>Auto mobile</th>
          <th>A12</th>
          <th>A8</th>
          <th>A0</th>
          <th>E10</th>
        </tr>
      </thead>
      <tbody className='text-right'>
        {ev.map((e, i) => <tr key={i}>
          <td>{e.team}</td>
          <td>{Math.round(team_sims[e.team]["activation"]["true"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["sustainability"]["true"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["auto_mobile"]["true"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["auto_charge"]["12"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["auto_charge"]["8"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["auto_charge"]["0"]/10)}%</td>
          <td>{Math.round(team_sims[e.team]["endgame"]["10"]/10)}%</td>
          </tr>)}
      </tbody>
    </table>
  )
}

export default EVTable;

