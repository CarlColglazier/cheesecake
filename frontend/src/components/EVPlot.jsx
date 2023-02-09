import React from 'react';
import * as Plot from "@observablehq/plot";
import { useRef, useEffect } from "react";


function quartile(points, counts, q) {
  const all_count = counts.reduce(function (a, b) { return a + b }, 0);
  let c = 0;
  for (let i = 0; i < points.length; i++) {
    c += counts[i];
    if (c > all_count*q) {
      return points[i];
    }
  }
}

function EVPlot({ data }) {
  const ref = useRef();

  let teams = data.map(function(e,i) { return e.team });

  let proc = data.flatMap(function(d) {
    return d.points.map(
      function(x, i) {
        return {"points":x, "team":d.team, "bcount":d["bcount"][i]};
      }
    )});

  const evmax = Math.max(...proc.map(a => a.points));

  useEffect(() => {
    const neworder = data.map(function (e) { 
      return { "team":e.team, "median":quartile(e.points, e.bcount, 0.5), "qlow":quartile(e.points, e.bcount, 0.1), "qhigh":quartile(e.points, e.bcount, 0.9) }
    }).sort(function(a, b) { return b.median - a.median });
    const neworderteams = neworder.map(function(e) { return e.team });
    const barChart = Plot.plot({
      marginLeft: 50,
      x: {
        axis: "top",
        domain: [0, evmax],
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
        Plot.dot(neworder, {x:"median", y:"team", fill:"#000"})
      ],
      width: 640,
      height: 15*teams.length,
    });
    ref.current.append(barChart);
    return () => barChart.remove();
  }, [data]);

  return (
    <div>
      <div ref={ref}></div>
    </div>
  );
}

export default EVPlot;
