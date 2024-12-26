package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	_, exists := streams[streamID]
	mu.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(ConvertJSONResponse("error", fmt.Sprintf("Stream %s not found", streamID), nil))
		return
	}

	job := Job{
		StreamID: streamID,
		Task:     "results",
		Payload:  nil,
	}

	select {
	case jobQueue <- job:
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write(ConvertJSONResponse("error", "Server is too busy. Please try again later.", nil))
		return
	}

	mu.RLock()
	ch := streams[streamID]
	mu.RUnlock()

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(ConvertJSONResponse("error", "Streaming not supported", nil))
		return
	}

	for msg := range ch {
		// TODO: send response to client in JSON payload
		_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
		if err != nil {
			log.Printf("Error flushing data to client for stream %s: %v", streamID, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(ConvertJSONResponse("error", "Internal Server error. Could not stream data", nil))
			return
		}

		flusher.Flush()

		var jsonResponse JSONResponse
		_ = json.Unmarshal([]byte(msg), &jsonResponse)
		if jsonResponse.Status == "error" {
			// Close stream due to error or timeout
			return
		}
	}
}

func ProcessStreamResults(job Job, workerID int) {
	streamID := job.StreamID

	mu.RLock()
	ch := streams[streamID]
	mu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokerAddrs,
		Topic:       streamID,
		StartOffset: kafka.LastOffset, // Start from the latest message
	})

	defer reader.Close()

	timer := time.NewTimer(ConsumerTimeout)
	defer timer.Stop()

	go func() {
		<-timer.C
		log.Printf("[Worker %d] Session timeout for stream %s", workerID, streamID)
		ch <- string(ConvertJSONResponse("error", "Timeout due to inactivity. Try again to start streaming", nil))
		cancel()
	}()

	// Consume Messages
	for {
		select {
		case <-ctx.Done():
			// Stop consumer after session end due to inactivity
			return

		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("[Worker %d] Error consuming message from stream %s", workerID, streamID)
				ch <- string(ConvertJSONResponse("error", "Error consuming message", nil))
			}

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(ConsumerTimeout)

			processedData := processMessage(string(msg.Value))
			ch <- string(ConvertJSONResponse("success", "Message consumed successfully", processedData))
			log.Printf("[Worker %d] Processed message for stream %s", workerID, streamID)

		}
	}

}

func processMessage(data string) string {
	// TODO: add a dynamic process time simulation (add more complex processing house temperature monitering)
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
