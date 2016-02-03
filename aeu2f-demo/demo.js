//
// Demonstration of U2F
// License: MIT
//

// Action is used to keep track of steps of the process.
function Action(name, data, style) {
  this.name = name
  this.data = data ? JSON.stringify(data, null, 2) : ''
  this.at = new Date().toLocaleString()
  this.css = Action.style_map[style] || ''
}

Action.style_map = {
  'fail': 'list-group-item-danger',
  'pass': 'list-group-item-success',
  'info': 'list-group-item-info'
}

var actions = ko.observableArray([new Action("Page Loaded", null, 'pass')])
var waiting_for_key = ko.observable(false)


Action.add = function(name, data, style) {
  actions.unshift(new Action(name, data, style))
}

var Model = {
  is_https: window.location.protocol === 'https:',
  supported: Boolean(window.u2f),
  actions: actions,
  waiting_for_key: waiting_for_key,

  onRegisterClick: function () {
    Action.add("Registration Challenge Requested", null, 'info')
    $.getJSON('/registerRequest')
      .done(function(req) {
        Action.add("Registration Challenge Received", req, 'info')
        waiting_for_key(true)
        u2f.register([req], [], afterTokenKeyPress, 20)
      })
      .fail(function(msg) {
        Action.add("Challenge Request Failed", msg, 'fail')
      });
  },

  onAuthenticateClick: function () {

  },
}


function afterTokenKeyPress(resp) {
  waiting_for_key(false)
  Action.add("Registration Challenge Ended", resp, 'info')
  $.post('/registerResponse', JSON.stringify(resp))
    .done(function() {
      Action.add("Registered", resp, 'pass')
    })
    .fail(function(msg) {
      Action.add("Registration Failed", msg, 'fail')
    });
}

function u2fSigned(resp) {
  $.post('/signResponse', JSON.stringify(resp)).done(function() {
    alert('Success');
  });
}

function sign() {
  $.getJSON('/signRequest').done(function(req) {
    u2f.sign([req], u2fSigned, 10);
  });
}


if (!window.u2f) {
  alert('Please install the U2F API.');
}


ko.applyBindings(Model)
