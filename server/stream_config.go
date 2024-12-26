package server

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

var (
	brokerAddrs = []string{
		"localhost:39092",
		"localhost:29092",
		"localhost:19092",
	}
	mu             sync.RWMutex
	streams        = make(map[string]chan string)
	streamContexts = make(map[string]context.CancelFunc)
	jobQueue       = make(chan Job, JobQueueSize)
	wg             sync.WaitGroup
)

type Job struct {
	StreamID string
	Task     string
	Payload  interface{}
	Result   chan []byte
}

type JSONResponse struct {
	Status  string
	Message string
	Data    interface{}
}

const (
	maxActiveStreams = 950  // Max active streams
	workerPoolSize   = 100  // Number of worker in the pools
	JobQueueSize     = 2000 // Maximum queued jobs

	// Timeout constant
	TopicCreationTimeOut = 5 * time.Second
	ProduderTimeOut      = 5 * time.Second
	ConusmerTimeOut      = 5 * time.Second
)

// Initialize workerpool
func InitializeWorkerPool() {
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go worker(i)
	}
}

// Populate workers with jobs
func worker(workerID int) {
	defer wg.Done()
	for job := range jobQueue {
		switch job.Task {
		case "start":
			ProcessStreamStart(job, workerID)
		case "send":
			ProcessStreamSend(job, workerID)
		case "results":

		case "end":

		}
	}
}

// Helper function to convert to Json response
func ConvertJSONResponse(status, message string, data interface{}) []byte {
	response := JSONResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	jsonResponse, _ := json.Marshal(response)
	return jsonResponse
}
