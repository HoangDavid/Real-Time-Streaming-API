package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
	"sync"
	"github.com/segmentio/kafka-go"
)


var (
	brokerAddr = "localhost:9092"
	results = make(map[string][]string)
	mu sync.Mutex
)

func StreamStart(w http.ResponseWriter, r *http.Request){
	//TODO: create topic if doesn't exists + start consumer goroutines

	streamID := mux.Vars(r)["stream_id"]

	err := createTopicIfNotExists(streamID)
	if err != nil{
		http.Error(w, fmt.Sprintf("Failed to create stream: %v", err), http.StatusInternalServerError)
		return
	}

	// start a consumer for a stream
	go ConsumeMessage(streamID)

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

	mu.Lock()
	data, exists := results[streamID]
	mu.Unlock()

	if !exists{
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	// TODO: Print the results
	mu.Lock()
	for i, result := range data {
        fmt.Printf("%d: %s\n", i+1, result)
    }
	mu.Unlock()
	
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
            // Topic already exists
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