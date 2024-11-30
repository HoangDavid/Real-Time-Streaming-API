package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
	"sync"
	"github.com/segmentio/kafka-go"
	"log"
)


var (
	brokerAddr = "localhost:9092"
	results = make(map[string][]string)
	mu sync.Mutex
	streams = make(map[string]chan string)
	wg sync.WaitGroup
)

// TODO: Able to handle support multiple streams
// JSON based request

// BUG!!!:
// 	Livelock if a second stream starts
//  Livelock on stream results

func StreamStart(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]

	err := createTopicIfNotExists(streamID)
	if err != nil{
		http.Error(w, fmt.Sprintf("Failed to create stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a channel for the stream
	mu.Lock()
	_, exists := streams[streamID]
	if !exists{
		streams[streamID] = make(chan string)
	}
	mu.Unlock()

	// Start a go routine for a stream
	wg.Add(1)
	go func() {
		defer wg.Done()
		ConsumeMessage(streamID)
	}()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Stream %s started successfully\n", streamID)))
}

func StreamSend(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]
	
	// Read data from request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
	}
	
	data := string(body)
	err = ProduceMessage(streamID, data)
	if err != nil{
		http.Error(w, fmt.Sprintf("Failed to send message to Kafka: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Data sent to stream %s\n", streamID)))
}

func StreamResults(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]

	// Set Server-side Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	mu.Lock()
	ch, exists := streams[streamID]
	mu.Unlock()

	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	// Stream the processed results to the client side
	for msg := range ch {
		_, err := fmt.Fprintf(w, "data: %s\n\n", msg) // SSE data format
		if err != nil{
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
        Topic: topic,
		NumPartitions: 3,
		ReplicationFactor: 1,
    }

	err = conn.CreateTopics(topicConfig)
    if err != nil {
        return err
    }

    return nil
}