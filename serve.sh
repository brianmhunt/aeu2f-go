#!/bin/sh

# We'll want to kill goapp when this exits.
# See e.g. http://stackoverflow.com/questions/360201
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# Start the AppEngine Server, in the background.
goapp serve ./aeu2f-demo &

# Connect with ngrok.
ngrok -proto https 8080
