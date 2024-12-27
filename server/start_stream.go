package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

// Create a channel for the stream
// TODO: write a bash script with wrk to load test the script somehow ??
// TODO: API authentication implementation
// TODO: write unit tests and integration tests
// DONE !!!

func StreamStart(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]
	result := make(chan []byte, 1)

	job := Job{
		StreamID: streamID,
		Task:     "start",
		Result:   result,
	}

	select {
	case jobQueue <- job:
		// TODO: add API authencation here
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

func ProcessStreamStart(job Job, workerID int) {
	streamID := job.StreamID

	mu.RLock()
	if len(streams) >= maxActiveStreams {
		// Server reached to max active streams
		log.Printf("[Worker %d] Reject start stream %s request. Max active connections reached!", workerID, streamID)
		job.Result <- ConvertJSONResponse("error", "Too many active connections. Please try again later", nil)
		return
	}

	_, exists := streams[streamID]
	mu.RUnlock()
	if exists {
		// Stream already exists
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Stream %s already exists", streamID), nil)
		return
	}

	err := createTopics(streamID)
	if err != nil {
		// Error creating  topics
		log.Printf("[Worker %d] %v\n", err, workerID)
		job.Result <- ConvertJSONResponse("error", fmt.Sprintf("Fail to create stream %s. Please try again", streamID), nil)
		return

	}

	// Admit a new stream to the server
	mu.Lock()
	streams[streamID] = make(chan string)
	mu.Unlock()
	log.Printf("[Worker %d] Stream %s starts successfully\n", workerID, streamID)
	job.Result <- ConvertJSONResponse("success", fmt.Sprintf("Stream %s starts successfully!", streamID), nil)

}

func createTopics(topic string) error {
	var conn *kafka.Conn
	var err error

	// Timeout if creating topic takes too long
	ctx, cancel := context.WithTimeout(context.Background(), TopicCreationTimeOut)
	defer cancel()

addrloop:
	for _, addr := range brokerAddrs {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stream %s timeout while connecting to %s", topic, addr)
		default:
			conn, err = kafka.Dial("tcp", addr)
			if err == nil {
				break addrloop
			}
		}

	}

	if conn == nil {
		return fmt.Errorf("failed to connect to any brokers")
	}
	defer conn.Close()

	controller, _ := conn.Controller()
	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("stream %s failed to dial controller broker %s: %v", topic, controllerAddr, err)
	}
	defer controllerConn.Close()

	// Check if the topic exists
	partitions, err := controllerConn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("stream %s failed to read partitions from controller: %v", topic, err)
	}

	for _, p := range partitions {
		if p.Topic == topic {
			return nil
		}
	}

	// Create topic
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 3,
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("stream %s timeout while creating topic", topic)
	default:
		err = controllerConn.CreateTopics(topicConfig)
		if err != nil {
			return fmt.Errorf("failed to create topic: %v", err)
		}
	}

	return nil
}
