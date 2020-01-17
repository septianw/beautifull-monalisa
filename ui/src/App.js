import React from 'react';
import './App.css';
import Register from './Register'
import { Container, Row, Col, Button } from 'reactstrap';

import { HashRouter as Router, Route, Switch } from 'react-router-dom';

var wsconn;

function WebsocketConnect(url) {
  wsconn = new WebSocket(url);
  wsconn.onopen = function () {
    console.log('wsconn: Connection opened.');
  }

  return wsconn;
}

class Login extends React.Component {
  onSubmit() {
    console.log("You are trying to login.");
  }
  render() {
    return <div id="wrapperin">
      <Container id="wrapperout" fluid="sm">
        <Row id="registerWrapper">
          <Col id="registerBlock" sm="12" md={{ size: 6, offset: 3 }}>
            <h2>Login</h2>
            <form onSubmit={this.onSubmit.bind(this)}>

              <input type="text" name="username" className="form-control" placeholder="Username." />
              <input type="password" name="password" className="form-control" placeholder="*******" />
              <Button color="primary" className="btn-register" >Login</Button>
            </form>
          </Col>
        </Row>
      </Container>
    </div>
  }
}

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      apikey: "",
      url: "",
      apiEndpoint: "/api/v1",
      apiProtocol: "http"
    }
  }

  componentDidMount() {
    let c = WebsocketConnect(this.props.url);
    c.onmessage = function (evt) {
      this.setState({
        apikey: evt.data
      })
    }.bind(this)

    this.setState(function (state, props) {
      state.url = props.url;
      return state;
    });
  }

  render() {
    return <Router>
      <Switch>
        <Route exact={true} path="/">
          <Register s={this.state} />
        </Route>
        <Router exact={true} path="/#login">
          <Login />
        </Router>
      </Switch>
    </Router>
  }
}

export default App;
