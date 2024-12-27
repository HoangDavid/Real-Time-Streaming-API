#!/usr/bin/env bash

# delete_all_topics_except_schemas.sh
# -----------------------------------
# Deletes all Redpanda topics except the '_schemas' topic.
# No extra output is printed.

BROKER="localhost:9092"

# List all topics from Redpanda (run inside the 'redpanda-0' container)
TOPICS=$(docker exec redpanda-0 rpk topic list --brokers="$BROKER")

# Loop through each topic; if it's not '_schemas', delete it
for topic in $TOPICS; do
  if [ "$topic" != "_schemas" ]; then
    docker exec redpanda-0 rpk topic delete "$topic" --brokers="$BROKER" >/dev/null 2>&1
  fi
done