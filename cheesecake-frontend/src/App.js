import React, { Component } from 'react';
import {
  Button,
  ButtonGroup,
  Collapse,
  Container,
  Navbar,
  NavbarToggler,
  NavbarBrand,
  Nav,
  NavItem,
  NavLink,
  Progress
} from 'reactstrap';
import socketIOClient from "socket.io-client";
import {
  BrowserRouter as Router,
  Route,
  Link
} from "react-router-dom";

const BASE_URL = `http://localhost:5000`;
const socket = socketIOClient(BASE_URL);

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
        {this.state.teams.map((team) => <li key={team.key}>{team.nickname}</li>)}
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
        <p>Click a button to run a task.</p>
        <ButtonGroup>
          <Button onClick={() => {
              socket.emit("teams");
              socket.on("teams", data => {
                this.setState({
                  "progress": data * 100
                });
              });
            }}>Teams</Button>
          <Button onClick={() => {
              socket.emit("districts");
            }}>Districts</Button>
          <Button onClick={() => {
              socket.emit("events");
              socket.on("events", data => {
                this.setState({
                  "progress": data * 100
                });
              });
            }}>Events</Button>
        </ButtonGroup>
        {this.state.progress > 0 && this.state.progress < 100 ? (
          <Progress value={this.state.progress} />
        ) : (
          <div></div>
        )}
      </div>
    );
  }
}

/*
const About = () => (
  <div>
    <h2>About</h2>
  </div>
);
*/

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
    socket.on("teams", data => {
      console.log(data);
    });
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
                  <NavItem>
                    <NavLink tag={Link} to="/teams">Teams</NavLink>
                  </NavItem>
                  <NavItem>
                    <NavLink tag={Link} to="/districts">Districts</NavLink>
                  </NavItem>
                </Nav>
              </Collapse>
            </Navbar>
          </header>
          <Container>
            <Route exact path="/" component={Home} />
            <Route path="/teams" component={Teams} />
            <Route path="/districts" component={Districts} />
          </Container>
        </div>
      </Router>
    );
  }
}

export default App;
