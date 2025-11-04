package consumer

import (
	"log"
	"os"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	consumerInstance *KafkaConsumer
	once             sync.Once
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
}

func NewConsumer(consumerGroup string, topics []string) (*KafkaConsumer, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               broker,
		"group.id":                        consumerGroup,
		"auto.offset.reset":               "earliest",
		"enable.auto.commit":              false,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
	})
	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Kafka consumer subscribed to: %v", topics)
	return &KafkaConsumer{consumer: c}, nil
}

func (kc *KafkaConsumer) Close() {
	kc.consumer.Close()
	log.Println("🧹 Kafka consumer closed.")
}
