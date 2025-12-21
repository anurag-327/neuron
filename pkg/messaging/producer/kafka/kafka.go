package kafkaProducer

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/anurag-327/neuron/pkg/messaging"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	addr   string
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
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
			Async:        true,
		},
		addr: broker,
	}

	log.Println("Kafka producer initialized.")
	return kp, nil
}

func (kp *Producer) Publish(topic string, key string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := kp.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Key:   []byte(key),
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

func (kp *Producer) Health() error {
	conn, err := kafka.Dial("tcp", kp.addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
