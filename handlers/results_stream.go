package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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
