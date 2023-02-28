import React from 'react';
import { EventDataType, MatchDataType } from '../types';

const MatchTable: React.FC<EventDataType> = ({ ev, matches, team_sims, predictions }) => {
    function prediction_entry(e: MatchDataType) {
        if ('predictions' in e) {
            const correct = (
                (e.predictions>0.5 && e.red_score>e.blue_score) ||
                (e.predictions<0.5 && e.blue_score>e.red_score)
            );
            const text_class = "text-right p-4 " + (correct ? '' : 'line-through');
            return (
                <td className={text_class} >
                    {Math.round(100*e.predictions)}%
                </td>
            )
        } else {
            return <td></td>
        }
    }

    function get_prediction(key: string, m: MatchDataType) {
        if (predictions === undefined) {
            return 0.5
        }
        return predictions[m.key][key]
    }

    return (
        <table className="table-auto w-full text-sm border">
            <thead className="border">
                <tr>
                    <th className='p-2'>Match</th>
                    <th>Red</th>
                    <th>Blue</th>
                    <th colSpan={2}>Scores</th>
                    <th>Activation</th>
                    <th>Sustainability</th>
                </tr>
            </thead>
            <tbody>
            {matches.map((e, i) => (
                <tr key={i} className="border">
                    <td className="text-right p-4">{e.match_number} {e.key}</td>
                    <td className="text-center p-4">{e.red.map((j: number[], k: number) => <span key={k} className='p-2'>{j}</span>)}</td>
                    <td className="text-center p-4">{e.blue.map((j: number[], k: number) => <span key={k} className='p-2'>{j}</span>)}</td>
                    <td className="text-right p-4">{e.red_score} ({Math.round(100*get_prediction("red_win", e))}%)</td>
                    <td className="text-right p-4">{e.blue_score} ({Math.round(100*get_prediction("blue_win", e))}%)</td>
                    {/*prediction_entry(e)*/}
                    <td>{Math.round(100*get_prediction("red_activation", e))}% | {Math.round(100*get_prediction("blue_activation", e))}%</td>
                    <td>{Math.round(100*get_prediction("red_sustainability", e))}% | {Math.round(100*get_prediction("blue_sustainability", e))}%</td>
                </tr>
            ))}
            </tbody>
        </table>
    )
}

export default MatchTable;
