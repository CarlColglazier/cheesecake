import React from 'react';
import * as Plot from "@observablehq/plot";
import { useRef, useEffect, useState } from "react";
import quartile from '../util';

function EVPlot({ data }) {
  const ref = useRef();
  const [conf, setConf] = useState(0.9);

  /*
  let proc = data.flatMap(function(d) {
    return d.points.map(
      function(x, i) {
        return {"points":x, "team":d.team, "bcount":d["bcount"][i]};
      }
    )});
  */

  const neworder = data.map(function (e) {

    let dict = {};
    for (let i = 0; i < e.points.length; i++) {
      dict[e.points[i]] = e.bcount[i];
    }

    const sortPoints = [...e.points].sort((a, b) => a - b);
    const sortCount = sortPoints.map(x => dict[x]);
    return { 
      "team":e.team,
      "sum": e.points.reduce((a, b) => a + b, 0),
      "median":quartile(sortPoints, sortCount, 0.5),
      "qlow":quartile(sortPoints, sortCount, 0.5-(conf/2)),
      "qhigh":quartile(sortPoints, sortCount, 0.5+(conf/2))
    }
  }).sort((a, b) => { return b.median - a.median });
  const neworderteams = neworder.map((e) => e.team);
  const evmax = Math.max(...neworder.map(a => a.qhigh));
  const evmin = Math.min(Math.min(...neworder.map(a => a.qlow)), 0.0);

  useEffect(() => {
    const barChart = Plot.plot({
      marginLeft: 50,
      x: {
        axis: "top",
        domain: [evmin, evmax],
        label: "Expected point above replacement per match"
      },
      y: {
        domain: neworderteams,
        label: "Teams"
      },
      color: {
        scheme: "YlGnBu"
      },
      marks: [
        //Plot.barX(proc, Plot.binX({fillOpacity: "proportion"}, {x:"points", y:"team", fillOpacity:"bcount"})),
        //Plot.tickX(proc, {x:"points", y:"team", strokeOpacity:"bcount"}),
        Plot.ruleY(neworder, {x1: "qlow", x2: "qhigh", y:"team"}),
        Plot.dot(neworder, {x:"median", y:"team", fill:"#000", title:"median"})
      ],
      //width: 640,
      height: 15*neworder.length,
    });
    ref.current.append(barChart);
    return () => barChart.remove();
  }, [neworder, conf]);

  return (
    <div>
      <div ref={ref}></div>
      <div className="slider">
        <label htmlFor="fader">Uncertainty interval {Math.round(100*conf)}%</label>
        <input type="range" min="50" list="intervals" value={Math.round(conf*100)} onChange={e => setConf(e.target.value/100)}/>
        <datalist id="intervals">
          <option>50</option>
          <option>75</option>
          <option>90</option>
          <option>99</option>
        </datalist>
      </div>
    </div>
  );
}

export default EVPlot;