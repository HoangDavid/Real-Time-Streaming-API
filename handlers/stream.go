package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

var (
	brokerAddr = "localhost:9092"
	// results = make(map[string][]string)
	mu      sync.RWMutex
	streams = make(map[string]chan string)
	wg      sync.WaitGroup
)

func StreamStart(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	err := createTopicIfNotExists(streamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a channel for the stream
	mu.Lock()
	_, exists := streams[streamID]
	if !exists {
		streams[streamID] = make(chan string)
	}
	mu.Unlock()

	// Start a go routine for a stream
	if !exists {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(streams[streamID])
			ConsumeMessage(streamID)
		}()

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("Stream %s started successfully\n", streamID)))
	} else {
		w.Write([]byte(fmt.Sprintf("Stream %s is already started\n", streamID)))
	}
}

func StreamSend(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	// Read data from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
	}

	data := string(body)
	err = ProduceMessage(streamID, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message to Kafka: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Data sent to stream %s\n", streamID)))
}

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

func createTopicIfNotExists(topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return err
	}

	for _, p := range partitions {
		if p.Topic == topic {
			return nil
		}
	}
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = conn.CreateTopics(topicConfig)
	if err != nil {
		return err
	}

	return nil
}

/**
TODO:
- Able to handle 1000 concurrent users
	- Load-balancing between many broker addresses

- Current user flow:
	- Create topic when connect
	- Delete topic when disconnect
	- Resource queueing when there are too many connections

- JSON Based request and response
- Add a end stream endpoint to release the resources (if idle, temporially release it??)
- Change the processing function for real-time updates (monitoring house temperature)
- Write a script to simulate client point of view (1 client to 1000 clients at least)
- Build a CLI tool for funzy (maybe)
*/
