package handlers


import (
	"github.com/segmentio/kafka-go"
	"context"
	"strings"
	"fmt"
	"log"
)

// TODO: Producer and Consumer should be able to handle multiple streams simultaneously

func ProduceMessage(streamID string, data string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID,
	})
	fmt.Printf("Producing message to topic %s: %s\n", streamID, data)
	
	// Release resource by kafka writer after the function is done
	defer writer.Close()

	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key: []byte(streamID),
			Value: []byte(data),
		},
	)

	if err != nil {
		return err
	}

	return nil
}


func ConsumeMessage(streamID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID, 
		StartOffset: kafka.LastOffset,
	})

	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil{
			log.Printf("Error reading message from stream %s: %v", streamID, err)
			return
		}
		
		processedData := ProcessMessage(string(msg.Value))

		// Add the processed results a global variable for simplicity
		mu.Lock()
		results[streamID] = append(results[streamID], processedData)
		mu.Unlock()

		// Send processed results to channels
		mu.Lock()
		ch, exists := streams[streamID]
		if exists {
			ch<-processedData
		}
		mu.Unlock()

		log.Printf("Processed message from stream %s: %s", streamID, processedData)
	}
}

func ProcessMessage(data string) string {
	return strings.ToUpper(data)
}

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092

//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create <topic_name>

//To check all the topics created in Kafka
// 	docker exec -it redpanda rpk topic list --brokers=localhost:9092

//To delete topics created:
//	docker exec -it redpanda rpk topic delete <topic_name> --brokers=localhost:9092