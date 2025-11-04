package producer

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	producerInstance *Producer
	once             sync.Once
)

// KafkaProducer wraps the confluent producer
type Producer struct {
	producer *kafka.Producer
}

// NewProducer creates a new Kafka producer instance
func NewProducer() (*Producer, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"acks":              "all",
	})
	if err != nil {
		return nil, err
	}

	kp := &Producer{producer: p}

	// Start background goroutine for delivery reports
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("❌ Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("✅ Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	log.Println("✅ Kafka producer initialized.")
	return kp, nil
}

// GetProducer returns the global Kafka producer instance, creating it if necessary
func GetProducer() *Producer {
	once.Do(func() {
		p, err := NewProducer()
		if err != nil {
			log.Fatalf("Failed to initialize Kafka producer: %v", err)
		}
		producerInstance = p
	})
	return producerInstance
}

func (kp *Producer) Produce(topic string, data []byte) error {
	deliveryChan := make(chan kafka.Event, 1)
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(kafka.PartitionAny),
		},
		Value: data,
	}
	err := kp.producer.Produce(msg, deliveryChan)
	if err != nil {
		return err
	}

	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			log.Printf("❌ Delivery failed: %v\n", m.TopicPartition.Error)
			close(deliveryChan)
			return m.TopicPartition.Error
		}
		log.Printf("✅ Delivered message to %v [%d] @ offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		close(deliveryChan)
	case <-time.After(5 * time.Second):
		// Timeout if Kafka doesn't respond
		close(deliveryChan)
		return fmt.Errorf("delivery timeout")
	}

	return nil
}

// Close gracefully flushes and closes the producer
func (kp *Producer) Close() {
	kp.producer.Flush(5000)
	kp.producer.Close()
	log.Println("🧹 Kafka producer closed.")
}
