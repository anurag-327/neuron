package kafkaConsumer

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
	consumerInstance messaging.Subscriber
	once             sync.Once
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewConsumer(consumerGroup string, topic string) (messaging.Subscriber, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		log.Println("KAFKA_BROKER not set, using default localhost:9092")
		broker = "localhost:9092"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     consumerGroup,
		Topic:       topic,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})

	log.Printf("✅ Kafka consumer initialized. Group: %s | Topic: %s", consumerGroup, topic)
	return &KafkaConsumer{reader: reader}, nil
}

func (kc *KafkaConsumer) Consume(ctx context.Context, handler func([]byte) error) {
	log.Println("🎧 Starting Kafka consumer loop...")
	for {
		m, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				log.Println("Consumer context canceled, stopping.")
				return
			}
			log.Printf("Kafka read error: %v", err)
			continue
		}

		if err := handler(m.Value); err != nil {
			log.Printf("Handler error: %v", err)
		}
	}
}

func (kc *KafkaConsumer) Close() {
	if kc.reader == nil {
		return
	}
	if err := kc.reader.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	} else {
		log.Println("Kafka consumer closed.")
	}
}
