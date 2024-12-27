package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

func StreamEnd(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	result := make(chan []byte)
	job := Job{
		StreamID: streamID,
		Task:     "end",
		Result:   result,
	}

	select {
	case jobQueue <- job:
		res := <-result
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(res))
	default:
		// Reject request when there are too many requests
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write(ConvertJSONResponse("error", "Server too busy =((. Please try again later", nil))
	}
}

func ProcessStreamEnd(job Job, workerID int) {
	streamID := job.StreamID

	mu.RLock()
	ch, exists := streams[streamID]
	mu.RUnlock()

	if !exists {
		log.Printf("[Worker %d] Stream %s not found", workerID, streamID)
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Stream %s not found", streamID), nil)
		return
	}

	// Remove active channel for stream
	mu.Lock()
	close(ch)
	delete(streams, streamID)
	mu.Unlock()

	// Delete Topic
	err := deleteTopic(streamID)
	if err != nil {
		log.Printf("[Worker %d] %v", workerID, err)
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Failed to start stream %s", streamID), nil)
	}

	// Force exit the goroutine for data streaming if active
	mu.RLock()
	cancel, exists := streamContexts[streamID]
	if exists {
		cancel()
	}
	mu.RUnlock()

	log.Printf("[Worker %d] Stream %s ended", workerID, streamID)
	job.Result <- ConvertJSONResponse("success", fmt.Sprintf("Stream %s disconnected successfully", streamID), nil)

}

func deleteTopic(stream_id string) error {
	var conn *kafka.Conn
	var err error

	for _, addr := range brokerAddrs {
		conn, err = kafka.Dial("tcp", addr)
		if err == nil {
			break
		}
	}

	if conn == nil {
		return fmt.Errorf("could not connect to any brokers")
	}
	defer conn.Close()

	controller, _ := conn.Controller()
	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("failed to dial controller broker %s: %w", controllerAddr, err)
	}
	defer controllerConn.Close()

	// Delete the topic
	err = controllerConn.DeleteTopics(stream_id)
	if err != nil {
		return fmt.Errorf("failed to delete topic %q: %w", stream_id, err)
	}

	log.Printf("Topic %s deleted successfully\n", stream_id)
	return nil
}
