package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

var (
	brokerAddrs = []string{
		"localhost:39092",
		"localhost:29092",
		"localhost:19092",
	}
	mu             sync.RWMutex
	streams        = make(map[string]chan string)
	streamContexts = make(map[string]context.CancelFunc)
	wg             sync.WaitGroup
)

const max_streams = 400

func StreamStart(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	// Create a channel for the stream
	// TODO: add a topic creation limit (400) so to not exceed partition memory limit
	// TODO: add workerpool so to limit CPU usage (2000)
	// TODO: add a timeout for consumer when the user is not sending requests and not ending the connections
	// TODO: goroutines for producer as well
	// TODO: write a bash script with wrk to load test the script somehow ??
	// TODO: API authentication implementation
	// TODO: write unit tests and integration tests
	// DONE !!!
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()

	_, exists := streams[streamID]
	if !exists {

		if len(streams) < max_streams {
			time.Sleep(50 * time.Millisecond) // Avoid spawning too many processes
			createTopics(streamID)

			streams[streamID] = make(chan string)
			ctx, cancel := context.WithCancel(context.Background())
			streamContexts[streamID] = cancel

			wg.Add(1) // might be helpful when testing with client
			go func() {
				defer wg.Done()
				defer close(streams[streamID])
				ConsumeMessage(ctx, streamID)
			}()

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(fmt.Sprintf("Stream %s started successfully\n", streamID)))
			return

		} else {
			http.Error(w, "Maximum stream limit has reached", http.StatusTooManyRequests)
			return
		}

	} else {
		w.Write([]byte(fmt.Sprintf("Stream %s is already started\n", streamID)))
		return
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

func StreamEnd(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	// Lock to safely access shared resources
	mu.Lock()
	_, exists := streams[streamID]
	if exists {
		delete(streams, streamID)

		if cancel, ok := streamContexts[streamID]; ok {
			cancel() // Cancel the context to stop the consumer
			delete(streamContexts, streamID)
		}
	}
	mu.Unlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Stream %s not found", streamID), http.StatusNotFound)
		return
	}

	log.Printf("Stream %s ended successfully\n", streamID)

	// Optionally delete the Kafka topic associated with the stream
	err := deleteTopic(streamID)
	if err != nil {
		log.Printf("Failed to delete topic %s: %v\n", streamID, err)
		http.Error(w, fmt.Sprintf("Stream %s ended but failed to delete topic: %v", streamID, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Stream %s ended successfully\n", streamID)))
}

func createTopics(topic string) error {
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

	// Check if the topic exists
	partitions, err := controllerConn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions from controller: %w", err)
	}

	for _, p := range partitions {
		if p.Topic == topic {
			log.Printf("Topic %s already exists.\n", topic)
			return nil
		}
	}

	// Create topic
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 3,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic %q: %w", topic, err)
	}

	log.Printf("Topic %s created successfully with retention settings.\n", topic)
	return nil
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
