package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

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
