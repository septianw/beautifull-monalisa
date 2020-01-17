import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import 'bootstrap/dist/css/bootstrap.min.css';

var us;
if (window.location.toString().split("/")[2] === "localhost:3000") {
  us = 'ws://127.0.0.1:5987/ws';
} else {
  us = 'ws://' + window.location.toString().split("/")[2] + '/ws';
}
console.log(us);

ReactDOM.render(<App url={us} />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
