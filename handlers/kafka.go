package handlers


import (
	"github.com/segmentio/kafka-go"
	"context"
)

var (
	brokerAddr = "localhost:9092"
)

func ProduceMessage(streamID string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID,
	})
	
	// Release resource by kafka writer after the function is done
	defer writer.Close()

	message := "Hello from stream " + streamID
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key: []byte(streamID),
			Value: []byte(message),
		},
	)

	if err != nil {
		return err
	}

	return nil
}


func ConsumeMessage(streamID string) (string, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID, 
	})

	// Relased resource by kafka reader after the function is done
	defer reader.Close()

	// Check on this ??
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil{
			return "", err
		}

		if string(msg.Key) == streamID {
			return string(msg.Value) + "\n", nil
		}
	}
}

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092
//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create default
//To check all the topics created in Kafka
// 	docker exec -it redpanda rpk topic list --brokers=localhost:9092