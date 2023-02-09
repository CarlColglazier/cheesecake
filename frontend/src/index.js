import React, { useState, useEffect } from "react";
import ReactDOM from "react-dom/client";
import {
  createBrowserRouter,
  createRoutesFromElements,
  Route,
  RouterProvider,
  Outlet,
  Link
} from "react-router-dom";
import "./index.css";
import Event from "./routes/Event";
import reportWebVitals from './reportWebVitals';

function Home() {
  return (
    <div className="container mx-auto px-4">
      <h1 className="text-3xl font-bold underline"><Link to="/events">Cheesecake</Link></h1>
      <Outlet />
    </div>
  )
};

function EventList() {
  const [data, setData] = useState({ isLoaded: false, events: [] });
  useEffect(() => {
    const result = fetch(`/api/events.json`)
        .then(res => res.json())
        .then(
      (result) => {
        console.log(result);
        setData({
            isLoaded: true,
            events: result
        })
      },
      (error) => {
        console.log(error);
        setData({
            isLoaded: false,
        })
      }
    );
  }, []);

  if (!data.isLoaded) {
    return <h1>Loading...</h1>;
  } else {
    return (
      <ul>
        {data.events.map((e, i) => (
          <li key={i} ><Link to={"/event/" + e}>{e}</Link></li>
        ))}
      </ul>
    );
  }
}

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route path="/" element={<Home />}>
      <Route path="events/" element={<EventList />} />
      <Route path="event/:eventId" element={<Event />} />
    </Route>
      
  )
);

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
