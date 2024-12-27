# Real Time Streaming API with Kafka(Redpanda)

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

#### 1. API Design
- **Endpoint**: `POST /stream/start`
  - **Description**: Start a new data stream.
  - **Request Payload**: None
  - **Response**:
    ```json
    {
      "status": "success",
      "message": "Stream started successfully!"
    }
    ```

- **Endpoint**: `POST /stream/{stream_id}/send`
  - **Description**: Send chunks of data to the server.
  - **Request Payload**:
    ```json
    {
      "data": "Message content here"
    }
    ```
  - **Response**:
    ```json
    {
      "status": "success",
      "message": "Data sent to stream {stream_id}!"
    }
    ```

- **Endpoint**: `GET /stream/{stream_id}/results`
  - **Description**: Retrieve real-time results for the stream.
  - **Response**:
    ```json
    {
      "status": "success",
      "data": "Processed data stream content"
    }
    ```

- **Endpoint**: `POST /stream/{stream_id}/end`
  - **Description**: Ends a stream and cleans up resources.
  - **Request Payload**: None
  - **Response**:
    ```json
    {
      "status": "success",
      "message": "Stream {stream_id} disconnected successfully"
    }
    ```
#### 2. Kafka (Redpanda) Integration
- **Message Broker**: Redpanda (a Kafka-compatible platform) is used to handle data streams for high throughput and fault tolerance.
- **Stream Isolation**: Each `stream_id` corresponds to a unique Kafka topic or partition to simulate isolated clients.
- **Consumer Processing**: Kafka consumers process the incoming data in real-time, simulating operations such as transformation or analytics.

#### 3. High-Level Architecture Diagram

```plaintext
Client ----> [API Gateway] ----> [Kafka Producer]
  |                                     |
  |                                     v
  |                                [Kafka Topic]
  |                                     |
  v                                     v
[Client SSE] <---- [Kafka Consumer] <---- [Stream Processing]
```

## Demo

![Demo of Real-Time Streaming API](graphic/demo.gif)
This demo showcases the real-time streaming API handling one connection.
Server(left) and Client (right)

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
```

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

### Key Insights

- **Optimal Performance**: Best performance with **100-200 concurrent connections**.
- **High Connection Impact**: Latency increases and throughput decreases with >300 connections.
- **Fault Tolerance**: Kafka ensures resilience and data integrity under load.
- **Resource Bottlenecks**: API struggles with heavy traffic due to server-side constraints.

---

### Future Improvements

- **Load Balancing**: Distribute traffic using tools like NGINX or HAProxy.
- **Horizontal Scaling**: Deploy multiple API server instances with Kafka partitioning.
- **Kafka Tuning**: Optimize producer/consumer configurations and increase partitions.
- **Connection Pooling**: Enhance persistent connection handling for lower latency.
- **Backpressure**: Regulate request rate to prevent server overload.
