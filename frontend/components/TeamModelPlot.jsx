import React from 'react';
import * as Plot from "@observablehq/plot";
import { useRef, useEffect } from "react";
import quartile from '../util';

function EVPlot({ data, model }) {
  const ref = useRef();

  const d = data[model];
  console.log(d);

  const neworder = d.map(function (e) {
    return {
      "team": e.team.toString(),
      "mean": e.mean,
      "low": e.low,
      "high": e.high
    }
  }).sort(function(a, b) { return b.mean - a.mean });

  console.log(neworder);

  useEffect(() => {
    const barChart = Plot.plot({
      marginLeft: 50,
      x: {
        axis: "top",
      },
      y: {
        domain: neworder.map(function(e) { return e.team }),
        label: "Teams"
      },
      marks: [
        Plot.ruleX([0], {strokeOpacity: 0.1}),
        Plot.ruleY(neworder, {x1: "low", x2: "high", y:"team"})
      ],
      height: 15*neworder.length,
    });
    ref.current.append(barChart);
    return () => barChart.remove();
  }, [d]);

  return (
    <div>
      <div ref={ref}></div>
    </div>
  );
}

export default EVPlot;