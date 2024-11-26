package main

import (
	"log"
	"net/http"
	// "github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

func main() {
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
	stream_id := vars["stream_id"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream " + stream_id + " sent\n"))
}

func streamResults(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	stream_id := vars["stream_id"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Show stream results"))
}
// Testing:
// Run: -X POST http://localhost:8080/stream/start
//