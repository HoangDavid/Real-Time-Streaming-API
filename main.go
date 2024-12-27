package main

import (
	"log"
	"net/http"
	"real-time-api/server"

	"github.com/gorilla/mux"
)

func main() {
	server.InitializeWorkerPool()

	r := mux.NewRouter()

	// r.Use(server.APIKeyAuthMiddleware)
	r.HandleFunc("/stream/{stream_id}/start", server.StreamStart).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/send", server.StreamSend).Methods("POST")
	r.HandleFunc("/stream/{stream_id}/results", server.StreamResults).Methods("GET")
	r.HandleFunc("/stream/{stream_id}/end", server.StreamEnd).Methods("POST")

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
