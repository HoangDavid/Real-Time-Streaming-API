package handlers

import (
	"context"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
)

// TODO: Producer and Consumer should be able to handle multiple streams simultaneously

func ProduceMessage(streamID string, data string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokerAddrs,
		Topic:   streamID,
	})
	log.Printf("Producing message to topic %s: %s\n", streamID, data)

	// Release resource by kafka writer after the function is done
	defer writer.Close()

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(streamID),
			Value: []byte(data),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func ConsumeMessage(ctx context.Context, streamID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokerAddrs,
		Topic:       streamID,
		StartOffset: kafka.LastOffset,
	})

	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer for stream %s ended\n", streamID)
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from stream %s: %v", streamID, err)
				return
			}

			processedData := ProcessMessage(string(msg.Value))

			mu.RLock()
			ch, exists := streams[streamID]
			mu.RUnlock()

			if exists {
				ch <- processedData
			}

			log.Printf("Processed message from stream %s: %s", streamID, processedData)
		}
	}
}

func ProcessMessage(data string) string {

	return strings.ToUpper(data)
}

// To run cluster on Docker:
//  docker-compose up -d

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092

//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create <topic_name>

//To check all the topics created in Kafka
// 	docker exec -it redpanda rpk topic list --brokers=localhost:9092

//To delete topics created:
//	docker exec -it redpanda rpk topic delete <topic_name> --brokers=localhost:9092
