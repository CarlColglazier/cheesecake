import React from 'react';
import * as Plot from "@observablehq/plot";
import { useRef, useEffect, useState } from "react";
import quartile from '../util';

function EVPlot({ data }) {
  const ref = useRef();

  const [conf, setConf] = useState(0.9);

  console.log(conf, 1-conf);
  let teams = data.map(function(e,i) { return e.team });

  let proc = data.flatMap(function(d) {
    return d.points.map(
      function(x, i) {
        return {"points":x, "team":d.team, "bcount":d["bcount"][i]};
      }
    )});

  const neworder = data.map(function (e) { 
    return { 
      "team":e.team,
      "median":quartile(e.points, e.bcount, 0.5),
      "qlow":quartile(e.points, e.bcount, 0.5-(conf/2)),
      "qhigh":quartile(e.points, e.bcount, 0.5+(conf/2)) }
  }).sort(function(a, b) { return b.median - a.median });
  const neworderteams = neworder.map(function(e) { return e.team });
  const evmax = Math.max(...neworder.map(a => a.qhigh));
  const evmin = Math.min(Math.min(...neworder.map(a => a.qlow)), 0.0);

  useEffect(() => {
    const barChart = Plot.plot({
      marginLeft: 50,
      x: {
        axis: "top",
        domain: [evmin, evmax],
        label: "Expected point contribution per match"
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
      height: 15*teams.length,
    });
    ref.current.append(barChart);
    return () => barChart.remove();
  }, [data, conf]);

  return (
    <div>
      <div ref={ref}></div>
      <div class="slider">
        <label for="fader">Uncertainty interval {Math.round(100*conf)}%</label>
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