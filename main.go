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

// Testing:
// Run: -X POST http://localhost:8080/stream/start
//