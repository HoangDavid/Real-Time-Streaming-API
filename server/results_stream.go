package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func StreamResults(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	// Set Server-side Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	mu.RLock()
	ch, exists := streams[streamID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	log.Printf("StreamResults called for streamID: %s", streamID)

	// Stream the processed results to the client side
	for msg := range ch {
		log.Printf("Stream %s: %s\n", streamID, msg)
		_, err := fmt.Fprintf(w, "data: %s\n\n", msg) // SSE data format
		if err != nil {
			log.Printf("Client disconnected from stream %s\n", streamID)
			return
		}

		w.(http.Flusher).Flush() // Sending the data to client immediately
	}

}

func ConsumeMessage(ctx context.Context, streamID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokerAddrs,
		Topic:       streamID,
		StartOffset: kafka.LastOffset,
	})

	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer for stream %s ended\n", streamID)
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from stream %s: %v", streamID, err)
				return
			}

			processedData := ProcessMessage(string(msg.Value))

			mu.RLock()
			ch, exists := streams[streamID]
			mu.RUnlock()

			if exists {
				ch <- processedData
			}

			log.Printf("Processed message from stream %s: %s", streamID, processedData)
		}
	}
}

func ProcessMessage(data string) string {

	return strings.ToUpper(data)
}

// To run cluster on Docker:
//  docker-compose up -d

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092

//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create <topic_name>

//To check all the topics created in Kafka
// 	docker exec -it redpanda rpk topic list --brokers=localhost:9092

//To delete topics created:
//	docker exec -it redpanda rpk topic delete <topic_name> --brokers=localhost:9092
