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
  TabContent,
  TabPane,
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
const YEAR = 2019;
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

class Simulation extends Component {
  constructor() {
    super();
    this.state = {
      data: {}
    };
  }

  componentDidMount() {
    fetch(BASE_URL + `/api/simulate/` + this.props.eventkey)
      .then(results => {
        return results.json();
      }).then(data => {
        this.setState({
          data: data
        });
      });
  }

  render() {
    return (
      <Table>
        <thead>
          <tr><th>Team</th><th>Average wins</th></tr>
        </thead>
        <tbody>
          {
            Object.keys(this.state.data).map(
              (key, index) => (
                <tr key={index}><td>{key}</td><td>{this.state.data[key]["mean"]}</td></tr>
              )
            )
        }
        </tbody>
      </Table>
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
              <td style={{backgroundColor: predict_color(match.predictions.EloScorePredictor)}}>{predict(match.predictions.EloScorePredictor)}</td>
              <td>{match.winning_alliance}</td>
              <td>{predict(match.predictions.EloScorePredictor).toLowerCase().includes(match.winning_alliance) || predict(match.predictions.EloScorePredictor) === "Tossup" ? '' : 'x'}</td>
            </tr>
          ))}
        </tbody>
      </Table>
    );
  }
}

class Event extends Component {
  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.state = {
      activeTab: '1'
    };
  }

  toggle(tab) {
    if (this.state.activeTab !== tab) {
      this.setState({
        activeTab: tab
      });
    }
  }
  render() {
    return (
      <div>
        <h2>{this.props.match.params.key}</h2>
        <Nav tabs>
          <NavItem>
            <NavLink
              onClick={() => { this.toggle('1'); }}
              >Simulation</NavLink>
          </NavItem>
          <NavItem onClick={() => { this.toggle('2'); }}>
            <NavLink>Matches</NavLink>
          </NavItem>
        </Nav>
        <TabContent activeTab={this.state.activeTab}>
          <TabPane tabId="1">
            <Simulation eventkey={this.props.match.params.key}/>
          </TabPane>
          <TabPane tabId="2">
            <MatchTable eventkey={this.props.match.params.key}/>
          </TabPane>
        </TabContent>
      </div>
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
