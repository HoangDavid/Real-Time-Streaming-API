package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
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
	Payload  string
	w        http.ResponseWriter
}

const (
	maxActiveStreams = 950  // Max active streams
	workerPoolSize   = 1000 // Number of worker in the pools
	JobQueueSize     = 2000 // Maximum queued jobs
)

// TimeOut constant
const (
	TopicCreationTimeOut = 5 * time.Second
)

func InitializeWorkerPool() {
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(1)
		go worker(i)
	}
}

func worker(workerID int) {
	defer wg.Done()
	for job := range jobQueue {
		switch job.Task {
		case "start":
			ProcessStartStream(job, workerID)
		case "send":

		case "results":

		case "end":

		}
	}
}

// Create a channel for the stream
// TODO: add a topic creation limit (400) so to not exceed partition memory limit
// TODO: add workerpool so to limit CPU usage (2000)
// TODO: add a timeout for consumer when the user is not sending requests and not ending the connections
// TODO: add a timeout for start stream (ask the user to retry as well) and end stream
// TODO: goroutines for producer and start/end stream as well
// TODO: write a bash script with wrk to load test the script somehow ??
// TODO: API authentication implementation
// TODO: write unit tests and integration tests
// DONE !!!

func StreamStart(w http.ResponseWriter, r *http.Request) {
	streamID := mux.Vars(r)["stream_id"]

	job := Job{
		StreamID: streamID,
		Task:     "start",
		w:        w,
	}

	select {
	case jobQueue <- job:
		// TODO: add API authencation here

		w.WriteHeader(http.StatusAccepted)

		// Stream start request is add to the worker pool
		w.Write([]byte(fmt.Sprintf("Start stream %s request has been accepted!!\n", streamID)))

	default:
		// Reject request when there are too many requeust
		http.Error(w, "Server is too busy. Please try again later.\n", http.StatusTooManyRequests)
	}
}

func ProcessStartStream(job Job, workerID int) {
	streamID := job.StreamID

	mu.Lock()
	defer mu.Unlock()

	if len(streams) >= maxActiveStreams {
		// Server reached the max active connections (overloads kafka broker)
		log.Printf("[Worker %d] Server reached active connections\n", workerID)
		http.Error(job.w, "Server too busy. Please try again later\n", http.StatusTooManyRequests)
		return
	}

	_, exists := streams[streamID]
	if exists {
		// Stream already exists
		http.Error(job.w, fmt.Sprintf("Stream %s already started\n", streamID), http.StatusConflict) // BUG: superfluouse writer to client
		return
	}

	err := createTopics(streamID)
	if err != nil {
		// Error creating  topics
		log.Printf("[Worker %d] %v\n", err, workerID)
		http.Error(job.w, fmt.Sprintf("failed to create stream %s =((. Please try again later\n", streamID), http.StatusInternalServerError)
		return

	}

	// Admit a new stream to the server
	streams[streamID] = make(chan string)
	log.Printf("[Worker %d] Stream %s starts successfully\n", workerID, streamID)
	job.w.Write([]byte(fmt.Sprintf("Stream %s starts sucessfully\n", streamID)))

}

func createTopics(topic string) error {
	var conn *kafka.Conn
	var err error

	// Timeout if creating topic takes too long
	ctx, cancel := context.WithTimeout(context.Background(), TopicCreationTimeOut)
	defer cancel()

addrloop:
	for _, addr := range brokerAddrs {
		select {
		case <-ctx.Done():
			return fmt.Errorf("ftream %s timeout while connecting to %s", topic, addr)
		default:
			conn, err = kafka.Dial("tcp", addr)
			if err == nil {
				break addrloop
			}
		}

	}

	if conn == nil {
		return fmt.Errorf("failed to connect to any brokers")
	}
	defer conn.Close()

	controller, _ := conn.Controller()
	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("stream %s failed to dial controller broker %s: %v", topic, controllerAddr, err)
	}
	defer controllerConn.Close()

	// Check if the topic exists
	partitions, err := controllerConn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("stream %s failed to read partitions from controller: %v", topic, err)
	}

	for _, p := range partitions {
		if p.Topic == topic {
			return nil
		}
	}

	// Create topic
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 3,
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("stream %s timeout while creating topic", topic)
	default:
		err = controllerConn.CreateTopics(topicConfig)
		if err != nil {
			return fmt.Errorf("failed to create topic: %v", err)
		}
	}

	return nil
}
