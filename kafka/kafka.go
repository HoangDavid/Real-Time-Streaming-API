package kafka

import (
	"fmt"
	"context"
	"log"
	"time"
	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitKafka() {
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic: "default",
		Balancer: &kafka.LeastBytes{},
	})
}

func ProduceMessage(topic string, message string) error {
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key: []byte(fmt.Sprintf("Key-%d", time.Now().Unix())),
			Value: []byte(message),
		},
	)
	
	return err
}

func ConsumeMessage(topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic: "default",
		 GroupID: "consumer-group-1",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil{
			log.Printf("could not read message %v", err)
			return
		}

		fmt.Printf("Message: %s\n", string(m.Value))
	}
}

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092
//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create default 