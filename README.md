# Real Time Streaming API with Redpanda(Kafka)

A high-performance real-time streaming API built with Golang and Kafka (Redpanda). This API enables clients to start streams, send data, and retrieve results via a scalable and robust architecture. Designed for handling large-scale concurrent connections, this API is benchmarked and optimized for real-time performance.

## Architecture

### Design Overview

- **API Design**:
  - RESTful endpoints in Golang for initiating streams, sending data, and retrieving real-time results.
  - Efficient management of concurrent connections using Goroutines.

- **Kafka (Redpanda) Integration**:
  - Each `stream_id` maps to a unique Kafka topic or partition ensure user isolation
  - Redpanda ensures fault tolerance and scalability for message processing.
  - Kafka consumers process incoming messages in real time.

- **Processing**:
  - Incoming data chunks are pushed to Kafka topics.
  - Consumers simulate real-time processing, transforming data dynamically.
  - Results are streamed back to clients via Server-Sent Events (SSE).

### Components

1. **API Server**:
   - Built in Golang.
   - Manages stream lifecycle: start, send, retrieve results, and end.
   - Ensures fault tolerance with graceful handling of inactive streams.

2. **Kafka Cluster**:
   - Uses a Redpanda-backed Kafka cluster with 3 brokers.
   - Each stream gets its own topic or partition for data isolation.

3. **Fault Tolerance**:
   - Kafka’s replication ensures no data loss during broker failures.
   - Timeouts for inactive streams prevent resource leakage.

---

## API Design

### Endpoints

  - `POST /stream/start`: Initializes a new stream.
  - `POST /stream/{stream_id}/send`: Sends data chunks to the specified stream.
  - `GET /stream/{stream_id}/results`: Retrieves real-time results using Server-Sent Events (SSE).
  - `POST /stream/{stream_id}/end`: Ends a stream and cleans up resources.

## Testing

- **Tool**: [wrk](https://github.com/wg/wrk)
- **Duration**: 30 seconds per test
- **Threads**: 3
- **Concurrent Connections**: Varying (100, 200, 300, ..., 800)
- **API Endpoint**: `http://localhost:8080`
- **Script**: `benchmark.lua` for simulating traffic patterns

### Command
```bash
wrk -t3 -c<CONNECTIONS> -d30s -s test/benchmark.lua http://localhost:8080

| Active connections | Requests/Sec | Transfer/Sec | Avg Latency | Max Latency | Socket Errors (Connect/Read/Write/Timeout) | Total Requests | Total Data Read |
|-------------|--------------|--------------|-------------|-------------|-------------------------------------------|----------------|-----------------|
| 100         | 81274.48     | 13.25MB      | 13.15ms     | 1.59s       | 0/0/0/84                                 | 2443754        | 398.35MB        |
| 200         | 60725.90     | 9.93MB       | 26.57ms     | 1.96s       | 0/47/0/145                               | 1827753        | 298.90MB        |
| 300         | 55103.13     | 9.02MB       | 32.63ms     | 1.97s       | 0/179/0/208                              | 1658462        | 271.51MB        |
| 400         | 39182.14     | 6.42MB       | 70.12ms     | 2.00s       | 0/341/0/211                              | 1179250        | 193.19MB        |
| 500         | 37079.17     | 6.08MB       | 89.75ms     | 2.00s       | 0/302/64/358                              | 1115027        | 182.74MB        |
| 600         | 33486.18     | 5.49MB       | 75.79ms     | 2.00s       | 0/679/92/298                              | 1007270        | 165.12MB        |
| 700         | 13327.64     | 2.19MB       | 172.25ms    | 2.00s       | 0/1424/0/827                              | 401533         | 65.88MB         |
| 800         | 14255.99     | 2.34MB       | 163.93ms    | 2.00s       | 0/1816/2/361                              | 428815         | 70.36MB         |
