#!/usr/bin/env bash -m

# Start the AppEngine Server, in the background.
# We don't want to use goapp here because it creates more child
# processes that we'd need to clean up.
dev_appserver.py ./aeu2f-demo &
GOAPP_PID=$!

# Connect with ngrok.
ngrok -log stdout -proto https 8080 > grok.log &
NGROK_PID=$!

# Wait for ngrok to connect
sleep 2

# Tell the user what the tunnel is
echo
echo "                             ⭐️ ⭐️ ⭐️   HTTPS Tunnel details  ⭐️ ⭐️ ⭐️ "
cat grok.log | grep "Tunnel established"

# Echo the jobs
echo "GOAPP PID ${GOAPP_PID} / NGROK PID ${NGROK_PID}"
jobs

# We'll want to kill the kids when this exits.
# See e.g. http://stackoverflow.com/questions/360201
trap "trap - SIGTERM && kill -- $GOAPP_PID $NGROK_PID" SIGINT SIGTERM EXIT

# Wait for processes to finish
# See e.g. http://stackoverflow.com/questions/356100
wait $GOAPP_PID
wait $NGROK_PID
