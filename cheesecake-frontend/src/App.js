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
  ListGroup,
  ListGroupItem,
//  Progress,
  Row,
  Table
} from 'reactstrap';
//import socketIOClient from "socket.io-client";
import {
  BrowserRouter as Router,
  Route,
  Link
} from "react-router-dom";

var BASE_URL;
if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
  BASE_URL = `http://` + window.location.hostname + `:5000`;
} else {
  BASE_URL = `https://` + window.location.hostname;
}
const YEAR = 2018;
const S1 = 0.19146;
const S2 = 0.34134;
const S3 = 0.43319;
//const socket = socketIOClient(BASE_URL);

function predict(fl) {
  if (typeof fl !== "number") {
    return "-";
  }
  if (fl < 0.5 - S3) {
    return "Likely Blue";
  } else if (fl < 0.5 - S2) {
    return "Leans Blue";
  } else if (fl < 0.5 - S1) {
    return "Tilts Blue";
  } else if (fl <= 0.5 + S1) {
    return "Tossup";
  } else if (fl <= 0.5 + S2) {
    return "Tilts Red";
  } else if (fl <= 0.5 + S3) {
    return "Leans Red";
  } else {
    return "Likely Red";
  }
}

function predict_color(fl) {
  if (typeof fl !== "number") {
    return "#ccc";
  }
  if (fl < 0.5 - S3) {
    return "#ccf";
  } else if (fl < 0.5 - S2) {
    return "#ddf";
  } else if (fl < 0.5 - S1) {
    return "#eef";
  } else if (fl <= 0.5 + S1) {
    return "white";
  } else if (fl <= 0.5 + S2) {
    return "#fee";
  } else if (fl <= 0.5 + S3) {
    return "#fdd";
  } else {
    return "#fcc";
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
    fetch(BASE_URL + '/api/events/' + YEAR)
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
      <div>
        <h2>Events</h2>
        <ListGroup>
          {this.state.events.map((e) => (
            <ListGroupItem key={e.key}><Link to={'/event/' + e.key}>{e.name}</Link></ListGroupItem>
          ))}
        </ListGroup>
      </div>
    );
  }
}

class MatchTable extends Component {
  constructor() {
    super();
    this.state = {
      matches: []
    };
  }

  componentDidMount() {
    console.log(this.props.eventkey);
    fetch(BASE_URL + `/api/matches/` + this.props.eventkey)
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          matches: data
        });
      });
  }

  render() {
    return (
      <Table>
        <thead>
          <tr>
            <th>Match</th>
            <th colSpan="6">Teams</th>
            <th>Prediction</th>
            <th>Winner</th>
            <th>Correct?</th>
          </tr>
        </thead>
        <tbody>
          {this.state.matches.map((match) => (
            <tr key={match.key}>
              <td>{match.key}</td>
              {match.alliances.red.team_keys.map((key) => (
                <td key={key}>{key.substring(3)}</td>
              ))}
            {match.alliances.blue.team_keys.map((key) => (
              <td key={key}>{key.substring(3)}</td>
            ))}
              <td style={{backgroundColor: predict_color(match.prediction)}}>{predict(match.prediction)}</td>
              <td>{match.winning_alliance}</td>
              <td>{predict(match.prediction).toLowerCase().includes(match.winning_alliance) || predict(match.prediction) === "Tossup" ? '' : 'x'}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    );
  }
}

const Event = (key) => (
  <div>
    <h2>{key.match.params.key}</h2>
    <MatchTable eventkey={key.match.params.key}/>
  </div>
);

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
        <p>Welcome to Cheesecake Live!</p>
      </div>
    );
  }
}

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
              <NavbarBrand href="#">
                <NavLink tag={Link} to="/">Cheesecake Live</NavLink>
              </NavbarBrand>
              <NavbarToggler onClick={this.toggle} />
              <Collapse isOpen={this.state.isOpen} navbar>
                <Nav className="m1-auto" navbar>
                  <NavItem>
                    <NavLink tag={Link} to="/events">Events</NavLink>
                  </NavItem>
                </Nav>
              </Collapse>
            </Navbar>
          </header>
          <Container>
            <Row>
              <Route exact path="/" component={Home} />
              <Route path="/events" component={EventList} />
              <Route path="/event/:key" component={Event} />
            </Row>
          </Container>
        </div>
      </Router>
    );
  }
}

export default App;
