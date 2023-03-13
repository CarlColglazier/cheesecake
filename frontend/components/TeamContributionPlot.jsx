import React from 'react';
import * as Plot from "@observablehq/plot";
import { useRef, useEffect } from "react";
import quartile from '../util';

function EVPlot({ data, team }) {
  const ref = useRef();

  const to_array = (o) => {
    return Object.entries(o).map(([k, v], i) => {
        return Array(v).fill(parseInt(k));
    }).flat().sort((a, b) => 0.5 - Math.random());
  }

  const tT = to_array(data[team]["tele_countT"]);
  const tM = to_array(data[team]["tele_countM"]);
  const tB = to_array(data[team]["tele_countB"]);
  const aT = to_array(data[team]["auto_countT"]);
  const aM = to_array(data[team]["auto_countM"]);
  const aB = to_array(data[team]["auto_countB"]);
  const link_count = to_array(data[team]["link_count"]);
  const auto = to_array(data[team]["auto_charge"]);
  const endgame = to_array(data[team]["endgame"]);
  const d = tT.map((v, i) => {
    return {
        "team": team,
        "tele": 5*tT[i] + 4*tM[i] + 3*tB[i] + 5*link_count[i],
        "auto": 6*aT[i] + 5*aM[i] + 4*aB[i] + auto[i] + endgame[i],
    };
  });

  console.log(team);
  console.log(d);

  //Object.entries(data).map((k));

  useEffect(() => {
    const chart = Plot.plot({
      marginLeft: 50,
      color: {
        scheme: "YlGnBu"
      },
      x: {domain: [0,100]},
      y: {domain: [0,75]},
      marks: [
        Plot.density(d, {x: "tele", y: "auto", fill:"density"}),
      ],
      //width: 640,
      //height: 15*teams.length,
    });
    ref.current.append(chart);
    return () => chart.remove();
  }, [d]);

  return (
    <div>
      <div ref={ref}></div>
    </div>
  );
}

export default EVPlot;