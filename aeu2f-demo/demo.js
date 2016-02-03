//
// Demonstration of U2F
// License: MIT
//

// Action is used to keep track of steps of the process.  It is only used
// for informational purposes and has no impact on the actual signing and
// registration.
function Action(name, data, style) {
  this.name = name
  this.data = data ? JSON.stringify(data, null, 2) : ''
  this.at = new Date().toLocaleString()
  this.css = Action.style_map[style] || ''
}

Action.style_map = {
  'fail': 'list-group-item-danger',
  'pass': 'list-group-item-success',
  'info': 'list-group-item-info',
  'warn': 'list-group-item-warning',
}

var actions = ko.observableArray([new Action("Page Loaded", null, 'pass')])
var waiting_for_key = ko.observable(false)
var is_communicating = ko.observable(false)

Action.add = function(name, data, style) {
  actions.unshift(new Action(name, data, style))
}

//
// ---  The signing steps are below ---
//


function on_fail(msg) {
  Action.add("Transmission failed", msg, 'fail')
  throw new Error("Transmission failed.")
}

/*
  request - Communicate with the server.

  Either GET a Challenge (to be completed by the U2F device), or POST
  the response completed by the U2F device.
 */
function request(type, url, data) {
  if (is_communicating()) { return }
  is_communicating(true)
  Action.add("[" + type + "] " + url, data, 'info')

  return $[type](url, data ? JSON.stringify(data) : undefined)
    .always(function () { is_communicating(false) })
    .fail(on_fail)
}



/*
  getU2FResponseToChallenge  -  Wait for the U2F interface to respond.

  Return a promise, so we keep to an easy-to-illustrate set of steps, below.
 */
function getU2FResponseToChallenge(kind, req) {
  Action.add("U2F Challenged to: " + kind, req, 'warn')
  waiting_for_key(true)
  var promise = $.Deferred()

  if (kind === 'register') {
    u2f.register([req], [], promise.resolve.bind(promise), 20)
  } else {  // kind is 'sign'
    u2f.sign([req], promise.resolve.bind(promise), 20)
  }

  promise.always(function () { waiting_for_key(false) })
  return promise
}


function sendChallengeResponse(url, resp) {
  // The resp could contain an errorCode, as per:
  // https://developers.yubico.com/U2F/Libraries/Client_error_codes.html
  if (resp.errorCode) {
    Action.add("U2F Failed.", resp, 'fail')
    throw new Error("U2F failed, error code: ", resp.errorCode)
  }
  return request("post", url, resp).fail(on_fail)
}


ko.applyBindings({
  is_https: window.location.protocol === 'https:',
  supported: Boolean(window.u2f),
  actions: actions,
  waiting_for_key: waiting_for_key,
  is_communicating: is_communicating,

  onRegisterClick: function () {
    request("getJSON", "/registerRequest")
      .then(getU2FResponseToChallenge.bind(null, 'register'))
      .then(sendChallengeResponse.bind(null, '/registerResponse'))
      .then(function () { Action.add("Registered", null, 'pass') })
  },

  onAuthenticateClick: function () {
    request("getJSON", "/signRequest")
      .then(getU2FResponseToChallenge.bind(null, 'sign'))
      .then(sendChallengeResponse.bind(null, '/signResponse'))
      .then(function () { Action.add("Signed", null, 'pass') })
  },
})
