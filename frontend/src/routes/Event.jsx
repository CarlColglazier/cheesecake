import React from 'react';
import { useEffect, useState } from "react";
import { useParams } from "react-router";
import EVPlot from '../components/EVPlot';
import MatchTable from '../components/MatchTable';


function Event() {
  let { eventId } = useParams();
  const [data, setData] = useState({ isLoaded: false, data: {} });
  useEffect(() => {
    const result = fetch(`/api/events/${eventId}.json`)
      .then(res => res.json())
      .then(
        (result) => {
          console.log(result);
          setData({
            isLoaded: true,
            data: result
          })
        },
        (error) => {
          setData({
              isLoaded: false,
          })
        }
    );
  }, []);

  if (!data.isLoaded) {
    return <h1>Loading...{ eventId }</h1>;
  } else {
    return (
      <div className="container mx-auto">
        <p className="text-sm">The current event is { eventId }</p>
        <EVPlot data={data.data.ev} className="" />
        <MatchTable data={data.data.matches} className="" />
      </div>
    )
  }
};

export default Event;
