import React from 'react';
import { EventDataType, MatchDataType } from '../types';

const MatchTable: React.FC<EventDataType> = ({ ev, matches, team_sims, predictions, schedule }) => {
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
                    <th colSpan={3}>Teams</th>
                    <th colSpan={2}>Scores</th>
                    <th colSpan={2}>Activation</th>
                    <th colSpan={2}>Sustainability</th>
                </tr>
            </thead>
            <tbody>
            {schedule.sort((a, b) => a.time - b.time).map((e, i) => (
            <>
                <tr key={i*2} className="border text-right p-2">
                    <td className="text-center">{e.key}</td>
                    {e.red_teams.map((j: number, k: number) => <td key={k} className='p-2 bg-red-50 text-center'>{j}</td>)}
                    <td className={e.blue_score < e.red_score ? 'font-semibold' : ''}>{e.red_score > 0 ? e.red_score : ""}</td>
                    <td className="text-right">{Math.round(100*get_prediction("red_win", e))}%</td>
                    <td></td>
                    <td className="text-right">{Math.round(100*get_prediction("red_activation", e))}%</td>
                    <td></td>
                    <td className="p-2">{Math.round(100*get_prediction("red_sustainability", e))}%</td>
                </tr>
                <tr key={i*2+1} className="border text-right p-2">
                    <td></td>
                    {e.blue_teams.map((j: number, k: number) => <td key={k} className='p-2 bg-blue-50 text-center'>{j}</td>)}
                    <td className={e.blue_score > e.red_score ? 'font-semibold' : ''}>{e.blue_score > 0 ? e.blue_score : ""}</td>
                    <td>{Math.round(100*get_prediction("blue_win", e))}%</td>
                    <td></td>
                    <td>{Math.round(100*get_prediction("blue_activation", e))}%</td>
                    <td></td>
                    <td className='p-2'>{Math.round(100*get_prediction("blue_sustainability", e))}%</td>
                </tr>
            </>
            ))}
            </tbody>
        </table>
    )
}

export default MatchTable;
