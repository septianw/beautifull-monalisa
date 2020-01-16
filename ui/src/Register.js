import React from 'react';
import { Container, Row, Col, Button, Tooltip } from 'reactstrap';
import { Dateselect } from './Htmlcomp';

/**
 * TODO:
 * 1. [x] GET VALUE FROM FORM AND SEND IT TO BACKEND
 * 2. [ ] PUT VALIDATION AROUND THE FORM
 * 3. [ ] PUT THE LOGIC OF DESIGN INTO WHOLE UI
 */

class LoginButton extends React.Component {
  render() {
    if (this.props.show) {
      return <Button color="primary" className="btn-login">Login</Button>
    } else {
      return null;
    }
  }
}

class Register extends React.Component {
  constructor(props) {
    super(props);
    this.refMobile = React.createRef();
    this.refEmail = React.createRef();
    this.refFirstname = React.createRef();
    this.refLastname = React.createRef();
    this.refGenderMale = React.createRef();
    this.refGenderFemale = React.createRef();
    this.refDOBday = React.createRef();
    this.refDOBmonth = React.createRef();
    this.refDOByear = React.createRef();
    this.refSubmitButton = React.createRef();

    this.state = {
      showLogin: false,
      tooltipMobileOpen: false,
      tooltipEmailOpen: false,
      tooltipFirstnameOpen: false,
      tooltipLastnameOpen: false,
      tooltipMobileMessage: "Please enter valid indonesian mobile phone number.",
      tooltipEmailMessage: "Seems like someone already used this email address to register here.",
      tooltipFirstnameMessage: "Firstname field cannot be empty.",
      tooltipLastnameMessage: "Lastname field cannot be empty."
    }
  }

  modifyState(modifiedState) {
    let originalState = JSON.parse(JSON.stringify(this.state));
    let modifiedOriginalState;

    for (var key in modifiedState) {
      if (originalState.hasOwnProperty(key) &&
          modifiedState.hasOwnProperty(key)) {
        originalState[key] = modifiedState[key]
      }
    }

    modifiedOriginalState = originalState;
    
    this.setState(function (state) {
      state = modifiedOriginalState;
      return state;
    });
  }

  validateMobile() {
    let numb = this.refMobile.current.value;
    // my case only cover cell phone number, this prefix can be extend.
    // @see https://en.wikipedia.org/wiki/List_of_mobile_telephone_prefixes_by_country
    const validPrefix = ["811", "812", "813", "814", "815", "816", "817", "818", "819", "838", "852", "853", "855", "856", "858", "859", "878", "896", "897", "898", "899"]
    let prefixStart = 0;
    let prefixEnd = 3;
    if (numb.substring(0, 1) === "0") {
      prefixStart = 1;
      prefixEnd = 4;
    }
    const input = numb.substring(prefixStart, prefixEnd);
    const len = numb.substring(prefixStart, numb.length).length;
    let out = false;

    let numprev = validPrefix.filter(function(c){
      return input === c;
    });
    if ((numprev.length > 0) && (len > 9) && (len < 12)) {
      out = true;
    }

    if (!out) {
      this.modifyState({
        tooltipMobileOpen: true,
        tooltipMobileMessage: "Please enter valid indonesian mobile phone number."
      })
    }

    return out;
  }

  validateEmail() {
    let email = this.refEmail.current.value;
    let regex = /[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?/;
    let isValid = regex.test(email);
    if (!isValid) {
      this.modifyState({
        tooltipEmailOpen: true,
        tooltipEmailMessage: "Pease enter valid email."
      })
    }
    return isValid;
  }

  onSubmit(e) {
    e.preventDefault();
    let value = {}
    let validMobile, validEmail;
    let validMobileMsg, validEmailMsg;

    // console.log(this.refMobile.current.value);
    // input should be like this:
    // {"mobile":"08123456750","email":"user@network.net","firstname":"Maya","lastname":"Lauren","date_of_birth":"20-03-1986","gender":""}
    value.mobile = this.refMobile.current.value;
    value.email = this.refEmail.current.value;
    value.firstname = this.refFirstname.current.value;
    value.lastname = this.refLastname.current.value;
    let dobraw = this.refDOBday.current.value + '-' + this.refDOBmonth.current.value + '-' + this.refDOByear.current.value;

    if (this.refGenderMale.current.checked) {
      value.gender = this.refGenderMale.current.value;
    }
    if (this.refGenderFemale.current.checked) {
      value.gender = this.refGenderFemale.current.value;
    }
    if (dobraw !== 'DEFAULT-DEFAULT-DEFAULT') {
      value.date_of_birth = dobraw;
    }

    validEmail = this.validateEmail();
    validMobile = this.validateMobile();

    if (validMobile && validEmail) {
      this.modifyState({
        tooltipEmailOpen: false,
        tooltipMobileOpen: false
      });
      console.log(value);
    //   this.send(value);
    }
  }

  send(formValue) {
    let url = new URL(this.props.s.url);
    let hit = async function (req) {
      let response = await fetch(req)
      let bodyText = await response.text();
      let bodyJson = {};
      let json = false;
      console.log(bodyText);
      if (response.headers.get('Content-Type').includes("application/json")) {
        try {
          bodyJson = JSON.parse(bodyText);
          json = true;
        } catch (e) {
          json = false;
          console.log(e);
        }
      }

      // if duplicate entry found.
      if (response.status === 409) {
        this.modifyState({
          tooltipMobileMessage: "Seems like someone already used this mobile number to register here.",
          tooltipEmailMessage: "Seems like someone already used this email to register here.",
          tooltipMobileOpen: true,
          tooltipEmailOpen: true
        })
      }

      // if required empty field found
      if (response.status === 400) {
        let t;
        if (json) {
          t = bodyJson.message;
        } else {
          t = bodyText;
        }
        if (t.includes("INPUT_VALIDATION_FAIL")) {
          let state = {};
          if (t.includes("Firstname")) {
            state.tooltipFirstnameOpen = true;
            state.tooltipFirstnameMessage = "Firstname field cannot be empty.";
          }

          if (t.includes("Lastname")) {
            state.tooltipLastnameOpen = true;
            state.tooltipLastnameMessage = "Lastname field cannot be empty.";
          }
          
          if (t.includes("Mobile")) {
            state.tooltipMobileOpen = true;
            state.tooltipMobileMessage = "Mobile field cannot be empty.";
          }

          if (t.includes("Email")) {
            state.tooltipEmailOpen = true;
            state.tooltipEmailMessage = "Email field cannot be empty.";
          }
          
          if (t.includes("does not validate as email")) {
            state.tooltipEmailOpen = true;
            state.tooltipEmailMessage = "Enter valid email address.";
          } else {
            this.setState({
              tooltipMobileOpen: !this.state.tooltipMobileOpen
            })
          }

          console.log(state);

          this.modifyState(state);
        }
      }

      if (response.status === 201) {
        this.setState({
          tooltipMobileOpen: false,
          tooltipEmailOpen: false,
          tooltipFirstnameOpen: false,
          tooltipLastnameOpen: false,
          showLogin: true
        })
      }

      // console.log(response.json());
      // console.log(response.text());
      console.log(response.status);
      console.log(response.ok);
    }.bind(this);
    url.protocol = this.props.s.apiProtocol;
    url.pathname = this.props.s.apiEndpoint + "/user/";
    console.log(url);
    console.log(this.props.s.url);
    console.log(url.toString());
    let head = new Headers();
    // head.append("Authorization", "Bearer c66f2005-b556-47e9-8086-328a354e6064");

    let req = new Request(url.toString(), {
      // mode: "no-cors",
      method: "POST",
      headers: head,
      body: JSON.stringify(formValue)
    });

    hit(req);
    console.log(req);
  }

  render() {
    return <div id="wrapperin">
      <Container id="wrapperout" fluid="sm">
        <Row id="registerWrapper">
          <Col id="registerBlock" sm="12" md={{ size: 6, offset: 3 }}>
            <h2>Registration</h2>
            <form onSubmit={this.onSubmit.bind(this)}>
              <Tooltip placement="top" isOpen={this.state.tooltipMobileOpen} autohide={true} target="mobileField" >{this.state.tooltipMobileMessage}</Tooltip>
              <input id="mobileField" type="number" required name="mobile" ref={this.refMobile} className="form-control" placeholder="Mobile number" />

              <Tooltip placement="top" isOpen={this.state.tooltipFirstnameOpen} autohide={true} target="firstnameField" >{this.state.tooltipFirstnameMessage}</Tooltip>
              <input id="firstnameField" type="text" required name="firstname" ref={this.refFirstname} className="form-control" placeholder="Firstname" />

              <Tooltip placement="top" isOpen={this.state.tooltipLastnameOpen} autohide={true} target="lastnameField" >{this.state.tooltipLastnameMessage}</Tooltip>
              <input id="lastnameField" type="text" required name="lastname" ref={this.refLastname} className="form-control" placeholder="Lastname" />

              <label htmlFor="dob-group">Date of Birth</label>
              <div id="dob-group" className="form-control nobgbd">
                <select name="month" ref={this.refDOBmonth} defaultValue={'DEFAULT'}>
                  <option value="DEFAULT" disabled>Month</option>
                  <option value="01">January</option>
                  <option value="02">February</option>
                  <option value="03">March</option>
                  <option value="04">April</option>
                  <option value="05">May</option>
                  <option value="06">June</option>
                  <option value="07">July</option>
                  <option value="08">August</option>
                  <option value="09">September</option>
                  <option value="10">October</option>
                  <option value="11">November</option>
                  <option value="12">December</option>
                </select>
                <select name="day" ref={this.refDOBday} defaultValue={'DEFAULT'}>
                  <option value="DEFAULT" disabled>Date</option>
                  <option value="01">01</option>
                  <option value="02">02</option>
                  <option value="03">03</option>
                  <option value="04">04</option>
                  <option value="05">05</option>
                  <option value="06">06</option>
                  <option value="07">07</option>
                  <option value="08">08</option>
                  <option value="09">09</option>
                  <option value="10">10</option>
                  <option value="11">11</option>
                  <option value="12">12</option>
                  <option value="13">13</option>
                  <option value="14">14</option>
                  <option value="15">15</option>
                  <option value="16">16</option>
                  <option value="17">17</option>
                  <option value="18">18</option>
                  <option value="19">19</option>
                  <option value="20">20</option>
                  <option value="21">21</option>
                  <option value="22">22</option>
                  <option value="23">23</option>
                  <option value="24">24</option>
                  <option value="25">25</option>
                  <option value="26">26</option>
                  <option value="27">27</option>
                  <option value="28">28</option>
                  <option value="29">29</option>
                  <option value="30">30</option>
                  <option value="31">31</option>
                </select>
                <select name="year" ref={this.refDOByear} defaultValue={'DEFAULT'}>
                  <option value="DEFAULT" disabled>Year</option>
                  <option value="1999">1999</option>
                  <option value="2000">2000</option>
                  <option value="2001">2001</option>
                  <option value="2002">2002</option>
                  <option value="2003">2003</option>
                  <option value="2004">2004</option>
                  <option value="2005">2005</option>
                  <option value="2006">2006</option>
                  <option value="2007">2007</option>
                  <option value="2008">2008</option>
                  <option value="2009">2009</option>
                  <option value="2010">2010</option>
                  <option value="2011">2011</option>
                  <option value="2012">2012</option>
                  <option value="2013">2013</option>
                  <option value="2014">2014</option>
                  <option value="2015">2015</option>
                </select>
                {/* January - 31 days
February - 28 days in a common year and 29 days in leap years
March - 31 days
April - 30 days
May - 31 days
June - 30 days
July - 31 days
August - 31 days
September - 30 days
October - 31 days
November - 30 days
December - 31 days */}
              </div>
              <div className="form-control genderSelect nobgbd" >
                <input id="genderSelectMale" type="radio" name="genderFieldMale" ref={this.refGenderMale} value="male" /> <label htmlFor="genderSelectMale" className="gender-form">Male</label>
                <input id="genderSelectFemale" type="radio" name="genderFieldFemale" ref={this.refGenderFemale} value="female" /> <label htmlFor="genderSelectFemale" className="gender-form">Female</label>
              </div>

              <Tooltip placement="top" isOpen={this.state.tooltipEmailOpen} autohide={true} target="emailField" >{this.state.tooltipEmailMessage}</Tooltip>
              <input id="emailField" type="text" required name="email" ref={this.refEmail} className="form-control" placeholder="Email" />

              <Button color="primary" className="btn-register" ref={this.refSubmitButton}>Register</Button>
            </form>
          </Col>
        </Row>
        <Row id="footerWrapper">
          <Col id="footerBlock" sm="12" md={{ size: 6, offset: 3 }}>
            <LoginButton show={this.state.showLogin} />
          </Col>
        </Row>
      </Container>
    </div>
  }
}

export default Register;