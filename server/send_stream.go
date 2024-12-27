package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func StreamSend(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	var requestData struct {
		Data string `json:"data"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.Data == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(ConvertJSONResponse("error", "Invalid request payload", nil))
		return
	}

	result := make(chan []byte, 1)

	job := Job{
		StreamID: streamID,
		Task:     "send",
		Payload:  requestData.Data,
		Result:   result,
	}

	select {
	case jobQueue <- job:
		res := <-result
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write(ConvertJSONResponse("error", "Server is too busy. Please try again later.", nil))
	}

}

func ProcessStreamSend(job Job, workerID int) {
	streamID := job.StreamID
	data, ok := job.Payload.(string)

	// Invalid request payload
	if !ok {
		job.Result <- ConvertJSONResponse("error", "Invalid payload for send stream", nil)
		return
	}

	mu.RLock()
	_, exists := streams[streamID]
	mu.RUnlock()

	// Stream not found
	if !exists {
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Stream %s not found", streamID), nil)
		return
	}

	// Error producing messaging
	err := ProduceMessage(streamID, data)
	if err != nil {
		log.Printf("[Worker %d] Stream %s failed to produce data to topic: %v", workerID, streamID, err)
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Failed to send data to stream %s", streamID), nil)
		return
	}

	log.Printf("[Worker %d] Data sent to stream %s\n", workerID, streamID)
	job.Result <- ConvertJSONResponse("success", fmt.Sprintf("Data sent to stream %s!", streamID), nil)
}

func ProduceMessage(streamID string, data string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ProduderTimeOut)
	defer cancel()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokerAddrs,
		Topic:        streamID,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	})

	defer writer.Close() // Release resource when done

	select {
	case <-ctx.Done():
		return fmt.Errorf("stream %s timeout while producing message", streamID)
	default:
		err := writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(streamID),
				Value: []byte(data),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to produce messsage: %v", err)
		}
	}

	return nil
}
