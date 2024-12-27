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
	workerPoolSize   = 1500 // Number of worker in the pools
	JobQueueSize     = 2000 // Maximum queued jobs

	// Timeout constant
	TopicCreationTimeOut = 5 * time.Second  // Prevent stall of topic creation
	ProduderTimeOut      = 5 * time.Second  // Prevent stall of producers
	ConsumerTimeout      = 20 * time.Second // Timeout for user inactivity
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
			ProcessStreamResults(job, workerID)
		case "end":
			ProcessStreamEnd(job, workerID)
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
	return append(jsonResponse, '\n')
}
