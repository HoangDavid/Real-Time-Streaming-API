package main

import (
	"log"
	"net/http"
	// "github.com/gorilla/websocket"
	"github.com/gorilla/mux"
	"real-time-api/kafka"
)

func main() {
	kafka.InitKafka()

	r := mux.NewRouter()
	r.HandleFunc("/stream/start", startStream).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/send", sendStream).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/results", streamResults).Methods("GET")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("could not start server: %s\n", err.Error())
    }
}

func startStream(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream started\n"))
	w.Write([]byte("Hello, I am ur first working stream in Golang\n"))
}

func sendStream(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
    streamID := vars["stream_id"]

    err := kafka.ProduceMessage(streamID, "Sample data chunk")
    if err != nil {
        http.Error(w, "Failed to send data to Kafka\n" + err.Error() + "\n", http.StatusInternalServerError)
        return
    }
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream sent!\n"))
}

func streamResults(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
    streamID := vars["stream_id"]

    go kafka.ConsumeMessage(streamID)

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Consuming stream results\n"))
}


// Testing:
// Run: -X POST http://localhost:8080/stream/start
//

// What is  a kafka topic? idk