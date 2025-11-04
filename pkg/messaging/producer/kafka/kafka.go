package kafkaProducer

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/anurag-327/neuron/pkg/messaging"
	"github.com/segmentio/kafka-go"
)

var (
	producerInstance *Producer
	once             sync.Once
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer() (messaging.Publisher, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		log.Println("KAFKA_BROKER not set, using default localhost:9092")
		broker = "localhost:9092"
	}

	kp := &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Balancer:     &kafka.Hash{},
			BatchSize:    100,
			BatchTimeout: 20 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
			Async:        true,
		},
	}

	log.Println("Kafka producer initialized.")
	return kp, nil
}

func (kp *Producer) Publish(topic string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := kp.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: data,
		},
	)

	if err != nil {
		log.Printf("Kafka delivery failed to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Kafka delivered message to topic %s", topic)
	return nil
}

func (kp *Producer) Close() {
	if kp == nil || kp.writer == nil {
		return
	}
	if err := kp.writer.Close(); err != nil {
		log.Printf("Error closing Kafka producer: %v", err)
	} else {
		log.Println("Kafka producer closed.")
	}
}
