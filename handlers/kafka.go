package handlers


import (
	"github.com/segmentio/kafka-go"
	"context"
	"strings"
	"fmt"
)

var (
	brokerAddr = "localhost:9092"
)

// TODO: Producer and Consumer should be able to handle multiple streams simultaneously

func ProduceMessage(streamID string, data string) error {
	err := createTopicIfNotExists(streamID)
	if err != nil{
		return err
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID,
	})
	fmt.Printf("Producing message to topic %s: %s\n", streamID, data)
	
	// Release resource by kafka writer after the function is done
	defer writer.Close()

	err = writer.WriteMessages(context.Background(),
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


func ConsumeMessage(streamID string) (string, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic: streamID, 
		StartOffset: kafka.LastOffset,
	})

	// Relased resource by kafka reader after the function is done
	defer reader.Close()

	// Consuming message and processing (BUG !!)
		// TODO: processing all the sent ones and then return
		// If all is processed, then return nothing

	msg, err := reader.ReadMessage(context.Background())
	if err != nil{
		return "", err
	}

	fmt.Printf("Consumed message from topic %s: %s\n", streamID, string(msg.Value))
    if string(msg.Key) == streamID {
		ret_val := ProcessMessage(string(msg.Value))
        return ret_val + "\n", nil
    }

	return "", nil
}

func ProcessMessage(data string) string {
	return strings.ToUpper(data)
}

func createTopicIfNotExists(topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	partitions, err := conn.ReadPartitions()
    if err != nil {
        return err
    }

    for _, p := range partitions {
        if p.Topic == topic {
            // Topic already exists
            return nil
        }
    }
	topicConfig := kafka.TopicConfig{
        Topic: topic,
		NumPartitions: 3,
		ReplicationFactor: 1,
    }

	err = conn.CreateTopics(topicConfig)
    if err != nil {
        return err
    }

    return nil
}

//To run redpanda on Docker:
//  docker run -d --name=redpanda -p 9092:9092 -p 9644:9644 docker.redpanda.com/vectorized/redpanda:latest redpanda start --overprovisioned --smp 1 --memory 1G --reserve-memory 0M --node-id 0 --check=false --kafka-addr PLAINTEXT://0.0.0.0:9092 --advertise-kafka-addr PLAINTEXT://localhost:9092

//To create a topic for Kafka:
//	docker exec -it redpanda rpk topic create default

//To check all the topics created in Kafka
// 	docker exec -it redpanda rpk topic list --brokers=localhost:9092

//To delete topics created:
//	docker exec -it redpanda rpk topic delete <topic_name> --brokers=localhost:9092