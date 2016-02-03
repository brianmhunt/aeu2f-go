
function register() {
  $.getJSON('/registerRequest').done(function(req) {
    console.log("Request Challenge", req)
    u2f.register([req], [], u2fRegistered, 2000)
  }).fail(console.error.bind(console));
}

function u2fRegistered(resp) {
  console.log("u2fRegistered.", resp)
  $.post('/registerResponse', JSON.stringify(resp)).done(function() {
    alert('Success');
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
