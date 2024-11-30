#!/bin/bash

# List all topics
topics=$(docker exec -it redpanda rpk topic list --brokers=localhost:9092 | awk 'NR>1 {print $1}')

# Delete each topic
for topic in $topics; do
  echo "Deleting topic: $topic"
  docker exec -it redpanda rpk topic delete $topic --brokers=localhost:9092
done

echo "All topics deleted."