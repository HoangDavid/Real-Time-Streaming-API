package main

import (
	"log"
	"net/http"
	"real-time-api/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/stream/{stream_id}/start", handlers.StreamStart).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/send", handlers.StreamSend).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/results", handlers.StreamResults).Methods("GET")
	r.HandleFunc("/stream/{stream_id}/end", handlers.StreamEnd).Methods("POST")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s\n", err.Error())
	}
}

//Testing:
// Run:
// curl -X POST http://localhost:8080/stream/123/start
// curl -X POST -d "Hi, my name is Hoang" http://localhost:8080/stream/123/send
// curl -X GET http://localhost:8080/stream/123/results
// curl -X POST http://localhost:8080/stream/123/end

//Note to self:
//	fuction has to be capitalized to be exported
//  to run docker, first kill process on monitor and then rerun the docker
