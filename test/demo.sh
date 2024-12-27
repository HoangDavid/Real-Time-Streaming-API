#!/bin/bash

# API Base URL
BASE_URL="http://localhost:8080"

# Stream ID to use
STREAM_ID=1

# Duration of the simulation in seconds
SIMULATION_DURATION=30

# Interval between sending messages (in seconds)
SEND_INTERVAL=1

# Message generator function (random string)
generate_message() {
  # Generate a random alphanumeric string of length 10-20
  local LENGTH=$((10 + RANDOM % 11))
  LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c $LENGTH
}

# Start the stream
start_stream() {
  echo "Starting stream $STREAM_ID..."
  curl -X POST "$BASE_URL/stream/$STREAM_ID/start" -s -o /dev/null
  echo "Stream started."
}

# Send random messages to the stream
send_messages() {
  echo "Sending messages to stream $STREAM_ID..."
  END_TIME=$((SECONDS + SIMULATION_DURATION))
  while [ $SECONDS -lt $END_TIME ]; do
    MESSAGE=$(generate_message)
    curl -X POST -H "Content-Type: application/json" \
      -d "{\"data\": \"$MESSAGE\"}" \
      "$BASE_URL/stream/$STREAM_ID/send" -s -o /dev/null
    sleep $SEND_INTERVAL
  done
  echo "Message sending completed."
}

# Fetch results from the stream
fetch_results() {
  echo "Fetching results from stream $STREAM_ID..."
  curl -N "$BASE_URL/stream/$STREAM_ID/results"
}

# Main script execution
main() {
  start_stream

  # Run sending messages and fetching results concurrently
  send_messages &
  fetch_results

  echo "Simulation completed."
}

main
