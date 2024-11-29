package handlers

import (
	"fmt"
	"io/ioutil"
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
	fmt.Fprintf(w, "Stream %s sent\n", streamID)
}

func StreamResults(w http.ResponseWriter, r *http.Request){
	streamID := mux.Vars(r)["stream_id"]

	msg, err := ConsumeMessage(streamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message to Kafka: %v", err), http.StatusInternalServerError)
		return
	}else if msg == "" {
		http.Error(w, fmt.Sprintf("Topic %s does not exist", streamID), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stream %s results\n", streamID)
	fmt.Fprintf(w, msg)
	
}