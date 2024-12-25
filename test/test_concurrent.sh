#!/bin/bash

# Function to start a stream, send a message, and get results
test_stream() {
  local stream_id=$1
  # echo "Starting stream $stream_id"
  curl -X POST http://localhost:8080/stream/$stream_id/start

  # echo "Sending message to stream $stream_id"
  # curl -X POST -d "Message for stream $stream_id" http://localhost:8080/stream/$stream_id/send

  # echo "Getting results for stream $stream_id"
  # curl -X GET http://localhost:8080/stream/$stream_id/results

}

# Test 10 concurrent streams
for i in {1..1000}; do
  test_stream $i &
  sleep 0.1
done

# Wait for all background processes to finish
wait

echo "All streams tested."