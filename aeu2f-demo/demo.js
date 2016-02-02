function checkExtension() {
  if (!window.u2f) {
    alert('Please install the Chrome U2F extension first.');
    return false;
  }
  return true;
}
function u2fRegistered(resp) {
  $.post('/registerResponse', JSON.stringify(resp)).done(function() {
    alert('Success');
  });
}
function register() {
  if (!checkExtension()) {
    return;
  }
  $.getJSON('/registerRequest').done(function(req) {
    u2f.register([req], [], u2fRegistered, 100)
  });
}
function u2fSigned(resp) {
  $.post('/signResponse', JSON.stringify(resp)).done(function() {
    alert('Success');
  });
}
function sign() {
  if (!checkExtension()) {
    return;
  }
  $.getJSON('/signRequest').done(function(req) {
    u2f.sign([req], u2fSigned, 10);
  });
}
