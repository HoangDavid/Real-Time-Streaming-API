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
