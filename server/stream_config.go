package server

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
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
	JobQueueSize     = 3000 // Maximum queued jobs

	// Timeout constant
	TopicCreationTimeOut = 5 * time.Second  // Prevent stall of topic creation
	ProduderTimeOut      = 5 * time.Second  // Prevent stall of producers
	ConsumerTimeout      = 20 * time.Second // Timeout for user inactivity
)

var workerPoolSize = runtime.NumCPU() * 20 // Number of worker in the pools
// Sample Valid API Keys (later replaced with database)
var validAPIKeys = map[string]bool{
	"your-api-key-1": true,
	"your-api-key-2": true,
}

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

func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" || !validAPIKeys[apiKey] {
			http.Error(w, string(ConvertJSONResponse("error", "Unauthorized: Invalid or missing API key", nil)), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
