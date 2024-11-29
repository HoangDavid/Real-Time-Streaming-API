package handlers

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

func StreamStart(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stream %s started\n", streamID)
}

func StreamSend(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]
	
	err := ProduceMessage(streamID)
	if err != nil{
		http.Error(w, fmt.Sprintf("Failed to send message to Kafka: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stream %s sent\n", streamID)
}

func StreamResults(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]

	msg, err := ConsumeMessage(streamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message to Kafka: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stream %s results\n", streamID)
	fmt.Fprintf(w, msg)
	
}