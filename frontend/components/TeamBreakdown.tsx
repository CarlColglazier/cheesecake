import React, { useState } from 'react';
import { EventDataType } from '../types';
import TeamContributionPlot from './TeamContributionPlot';
import TeamModelPlot from './TeamModelPlot';

const key_map: {[key: string]: string} = {
    'autoT': 'Auto (top)',
    'autoM': 'Auto (middle)',
    'autoB': 'Auto (bottom)',
    'teleT': 'Tele (top)',
    'teleM': 'Tele (middle)',
    'teleB': 'Tele (bottom)',
};

const key_description: {[key: string]: string} = {
    'autoT': 'Expected count of objects scored in the top row during autonomous',
    'autoM': 'Expected count of objects scored in the middle row during autonomous',
    'autoB': 'Expected count of objects scored in the bottom row during autonomous',
    'teleT': 'Expected logodds change in expectation an object will be scored in the top row during teleop',
    'teleM': 'Expected logodds change in expectation an object will be scored in the middle row during teleop',
    'teleB': 'Expected logodds change in expectation an object will be scored in the bottom row during teleop',
};

function get_human_key(key: string) {
    return key_map[key] || key;
}

const TeamBreakdown: React.FC<EventDataType>  = ({ev, matches, team_sims, predictions, schedule, model_summary}) => {
    const teams = Object.keys(team_sims)
    const [selectedOption, setSelectedOption] = useState(teams[Math.floor(teams.length*Math.random())]);

    const [selectedModel, setSelectedModel] = useState(model_summary ? Object.keys(model_summary)[0] : 'default');
    return (<>
        <select value={selectedModel} onChange={e => setSelectedModel(e.target.value)}>
            {model_summary ? Object.keys(model_summary).map((k) => {
                return <option key={k} value={k}>{get_human_key(k)}</option>
            }) : null}
        </select>
        <p>{key_description[selectedModel]}</p>
        <TeamModelPlot data={model_summary} model={selectedModel} />
    </>)
}

export default TeamBreakdown;