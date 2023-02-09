
import React from 'react';

function MatchTable({ data }) {

    function prediction_entry(e) {
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

    return (
        <table className="table-auto w-full text-sm border">
            <thead className="border">
                <tr>
                    <th className='p-2'>Match</th>
                    <th>Red</th>
                    <th>Blue</th>
                    <th colSpan={2}>Scores</th>
                    <th>Prediction</th>
                </tr>
            </thead>
            <tbody>
            {data.map((e, i) => (
                <tr key={i} className="border">
                    <td className="text-right p-4">{e.match_number}</td>
                    <td className="text-center p-4">{e.red.map((j, k) => <span key={k} className='p-2'>{j}</span>)}</td>
                    <td className="text-center p-4">{e.blue.map((j, k) => <span key={k} className='p-2'>{j}</span>)}</td>
                    <td className="text-right p-4">{e.red_score}</td>
                    <td className="text-right p-4">{e.blue_score}</td>
                    {prediction_entry(e)}
                </tr>
            ))}
            </tbody>
        </table>
    )
}

export default MatchTable;
