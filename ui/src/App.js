import React from 'react';
import './App.css';
import Register from './Register'

import { BrowserRouter as Router, Route } from 'react-router-dom';
import { stat } from 'fs';

var wsconn;

function WebsocketConnect(url) {
  wsconn = new WebSocket(url);
  wsconn.onopen = function () {
    console.log('wsconn: Connection opened.');
  }

  return wsconn;
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
      <Route path="/">
        <Register s={this.state} />
      </Route>
    </Router>
  }
}

export default App;
