import React, { useState } from 'react';
import { EventDataType } from '../types';
import TeamContributionPlot from './TeamContributionPlot';
import TeamModelPlot from './TeamModelPlot';

const TeamBreakdown: React.FC<EventDataType>  = ({ev, matches, team_sims, predictions, schedule, model_summary}) => {
    const teams = Object.keys(team_sims)
    const [selectedOption, setSelectedOption] = useState(teams[Math.floor(teams.length*Math.random())]);

    const [selectedModel, setSelectedModel] = useState(model_summary ? Object.keys(model_summary)[0] : 'default');
    return (<>

        <select value={selectedModel} onChange={e => setSelectedModel(e.target.value)}>
            {model_summary ? Object.keys(model_summary).map((k) => {
                return <option key={k} value={k}>{k}</option>
            }) : null}
        </select>

        <TeamModelPlot data={model_summary} model={selectedModel} />
        
        {/*
        <select value={selectedOption} onChange={e => setSelectedOption(e.target.value)}>
            {teams.map((k) => {
                return <option key={k} value={k}>{k}</option>
            })}
        </select>
        
        */}
    </>)
}

export default TeamBreakdown;