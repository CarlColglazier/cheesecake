import React, { Component } from 'react';
import {
  Collapse,
  Container,
  Navbar,
  NavbarToggler,
  NavbarBrand,
  Nav,
  NavItem,
  NavLink,
  Progress,
  Row,
  Table
} from 'reactstrap';
import socketIOClient from "socket.io-client";
import {
  BrowserRouter as Router,
  Route,
  Link
} from "react-router-dom";

const BASE_URL = `http://` + window.location.hostname + `:5000`;
const socket = socketIOClient(BASE_URL);

console.log(BASE_URL);

class TeamList extends Component {
  constructor() {
    super();
    this.state = {
      teams: []
    };
  }
  componentDidMount() {
    fetch(BASE_URL + '/api/teams/1')
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          'teams': data
        });
      });
  }
  render() {
    return (
      <ul>
        {this.state.teams.map((team) => (
          <li key={team.key}><Link to={'/team/' + team.key}>{team.nickname}</Link></li>
        ))}
      </ul>
    );
  }
}

class DistrictList extends Component {
  constructor() {
    super();
    this.state = {
      districts: []
    };
  }
  componentDidMount() {
    fetch(BASE_URL + '/api/districts')
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          'districts': data
        });
      });
  }
  render() {
    return (
      <ul>
        {this.state.districts.map((team) => <li key={team.key}>{team.display_name}</li>)}
      </ul>
    );
  }
}

class EventList extends Component {
  constructor() {
    super();
    this.state = {
      events: []
    };
  }
  componentDidMount() {
    fetch(BASE_URL + '/api/events')
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          events: data
        });
      });
  }
  render() {
    return (
      <ul>
        {this.state.events.map((e) => <li key={e.key}>{e.name}</li>)}
      </ul>
    );
  }
}

class EloList extends Component {
  constructor() {
    super();
    this.state = {
      teams: []
    };      
  }

  componentDidMount() {
    fetch(BASE_URL + `/api/elo`)
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          teams: data
        });
      });
  }

  render() {
    return (
      <Table>
        <tbody>
          {this.state.teams.map((e) => (
            <tr key={e.key}>
              <td><Link to={'/team/' + e.key}>{e.key}</Link></td>
              <td>{Math.round(e.score)}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    );
  }
}

class TeamHistory extends Component {
  constructor() {
    super();
    this.state = {
      matches: []
    };
  }

  componentDidMount() {
    console.log(this.props.team);
    fetch(BASE_URL + `/api/team/` + this.props.team + `/matches`)
      .then(results => {
        return results.json();
      }).then(data => {
        for (var match of data) {
          match["predictions"] = {
            "elo": 0.5
          };
        }
        this.setState({
          matches: data
        });
        fetch(BASE_URL + `/api/elo/` + this.props.team)
          .then(results => {
            return results.json();
          }).then(elo_data => {
            for (var match of data) {
              match["predictions"] = {
                "elo": elo_data[match["key"]]
              };
            }
            console.log(data);
            this.setState({
              matches: data
            });
          });
      });
  }

  predict(fl) {
    if (typeof fl !== "number") {
      return "-";
    }
    if (fl < 0.006) {
      return "Likely Blue";
    } else if (fl < 0.067) {
      return "Leans Blue";
    } else if (fl < 0.309) {
      return "Tilts Blue";
    } else if (fl < 0.6915) {
      return "Tossup";
    } else if (fl < 0.93319) {
      return "Tilts Red";
    } else if (fl < 0.99379) {
      return "Leans Red";
    } else {
      return "Likely Red";
    }
  }

  predict_color(fl) {
    if (typeof fl !== "number") {
      return "#ccc";
    }
    if (fl < 0.006) {
      return "#ccf";
    } else if (fl < 0.067) {
      return "#ddf";
    } else if (fl < 0.309) {
      return "#eef";
    } else if (fl < 0.691) {
      return "white";
    } else if (fl < 0.933) {
      return "#fee";
    } else if (fl < 0.994) {
      return "#fdd";
    } else {
      return "#fcc";
    }
  }

  render() {
    return (
      <Table>
        <tbody>
          <tr><th>Match</th><th colSpan="6">Teams</th><th>Red</th><th>Blue</th><th>Prediction</th><th>Correct?</th></tr>
          {this.state.matches.map((match) => (
            <tr key={match.key}>
              <td>{match.comp_level} {match.match_number}</td>
              {match.alliances.red.team_keys.map((key) => (
                <td key={key}>{key.substring(3)}</td>
              ))}
            {match.alliances.blue.team_keys.map((key) => (
              <td key={key}>{key.substring(3)}</td>
            ))}
              <td className={match.winning_alliance === "red" ? 'winner' : 'loser'}>
              {match.alliances.red.score}
            </td>
              <td className={match.winning_alliance === "blue" ? 'winner' : 'loser'}>
              {match.alliances.blue.score}
            </td>
              <td style={{backgroundColor: this.predict_color(match.predictions.elo)}}>
                {this.predict(match.predictions.elo)}
            </td>
              <td>
              {this.predict(match.predictions.elo).toLowerCase().includes(match.winning_alliance) || this.predict(match.predictions.elo) === "Tossup" ? '' : 'x'}
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
    );
  }
}

class Home extends Component {
  constructor() {
    super();
    this.state = {
      progress: 0
    };
  }
  render() {
    return (
      <div>
        <h2>Home</h2>
      </div>
    );
  }
}

const Teams = () => (
  <div>
    <h2>Teams</h2>
    <TeamList/>
  </div>
);

const Districts = () => (
  <div>
    <h2>Districts</h2>
    <DistrictList/>
  </div>
);

const Events = () => (
  <div>
    <h2>Events</h2>
    <EventList/>
  </div>
);

const Elo = () => (
  <div>
    <h2>Rankings</h2>
    <EloList/>
  </div>
);

const TeamData = (url) => (
  <div>
    <h3>Team History - {url.match.params.key}</h3>
    <TeamHistory team={url.match.params.key}/>
  </div>
);

class App extends Component {
  constructor(props) {
    super(props);
    this.toggle = this.toggle.bind(this);
    this.state = {
      isOpen: false
    };
  }
  toggle = () => {
    this.setState({
      isOpen: !this.state.isOpen
    });
  }
  componentDidMount = () => {
    //
  }
  render() {
    return (
      <Router>
        <div className="App">
          <header className="App-header">
            <Navbar color="dark" dark expand="md">
              <NavbarBrand href="#">Cheesecake</NavbarBrand>
              <NavbarToggler onClick={this.toggle} />
              <Collapse isOpen={this.state.isOpen} navbar>
                <Nav className="m1-auto" navbar>
                  <NavItem>
                    <NavLink tag={Link} to="/">Home</NavLink>
                  </NavItem>
                  {/*
                  <NavItem>
                    <NavLink tag={Link} to="/teams">Teams</NavLink>
                  </NavItem>
                  <NavItem>
                    <NavLink tag={Link} to="/districts">Districts</NavLink>
                  </NavItem>
                  <NavItem>
                    <NavLink tag={Link} to="/events">Events</NavLink>
                  </NavItem>
                  */}
                  <NavItem>
                    <NavLink tag={Link} to="/elo">Rankings</NavLink>
                  </NavItem>
                </Nav>
              </Collapse>
            </Navbar>
          </header>
          <Container>
            <Row>
              <Route exact path="/" component={Home} />
              <Route path="/teams" component={Teams} />
              <Route path="/districts" component={Districts} />
              <Route path="/events" component={Events} />
              <Route path="/elo" component={Elo} />
              <Route path="/team/:key" component={TeamData} />
            </Row>
          </Container>
        </div>
      </Router>
    );
  }
}

export default App;
