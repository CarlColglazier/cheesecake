import React, { useState } from 'react';
import { EventDataType } from '../types';
import TeamContributionPlot from './TeamContributionPlot';

const TeamBreakdown: React.FC<EventDataType>  = ({ev, matches, team_sims, predictions, schedule}) => {
    const teams = Object.keys(team_sims)
    const [selectedOption, setSelectedOption] = useState(teams[Math.floor(teams.length*Math.random())]);
    return (<>
        <select value={selectedOption} onChange={e => setSelectedOption(e.target.value)}>
            {teams.map((k) => {
                return <option key={k} value={k}>{k}</option>
            })}
        </select>
        <TeamContributionPlot data={team_sims} team={selectedOption} />
    </>)
}

export default TeamBreakdown;