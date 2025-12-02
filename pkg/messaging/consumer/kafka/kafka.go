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
	topic  string
	group  string
}

func NewConsumer(consumerGroup string, topic string) (messaging.Subscriber, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		log.Println("⚠️ KAFKA_BROKER not set, using default localhost:9092")
		broker = "localhost:9092"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           []string{broker},
		GroupID:           consumerGroup,
		Topic:             topic,
		MinBytes:          1,
		MaxBytes:          10e6,
		MaxWait:           200 * time.Millisecond,
		StartOffset:       kafka.LastOffset,
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		ReadLagInterval:   -1,
	})

	log.Printf("✅ Kafka consumer initialized. Group: %s | Topic: %s", consumerGroup, topic)
	return &KafkaConsumer{reader: reader, topic: topic, group: consumerGroup}, nil
}

// Consume is the normal streaming consumer —
// internally calls ConsumeControlled with maxConcurrent = 0 (no limit).
func (kc *KafkaConsumer) Consume(ctx context.Context, handler func([]byte) error) {
	log.Printf("⚡ Starting unbounded consumer for topic=%s", kc.topic)
	kc.ConsumeControlled(ctx, handler, 0)
}

// ConsumeControlled reads messages with bounded concurrency.
// If maxConcurrent = 0 → behaves like normal Consume() (no limit).
func (kc *KafkaConsumer) ConsumeControlled(ctx context.Context, handler func([]byte) error, maxConcurrent int) {
	var sem chan struct{}
	if maxConcurrent > 0 {
		sem = make(chan struct{}, maxConcurrent)
		log.Printf("🚀 Controlled consumer started for topic=%s (limit=%d)", kc.topic, maxConcurrent)
	} else {
		log.Printf("⚡ Unbounded consumer started for topic=%s", kc.topic)
	}

	for {
		// Apply backpressure only if a limit is set
		if sem != nil {
			select {
			case sem <- struct{}{}: // Acquire slot
			case <-ctx.Done():
				log.Printf("🛑 Context canceled for topic=%s", kc.topic)
				return
			}
		}

		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			// Release slot if we acquired one
			if sem != nil {
				select {
				case <-sem:
				default:
				}
			}

			if err == context.Canceled {
				log.Printf("🛑 Consumer context canceled for topic=%s", kc.topic)
				return
			}
			log.Printf("⚠️ Kafka read error [%s]: %v", kc.topic, err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process message concurrently (bounded by sem)
		go func(value []byte) {
			defer func() {
				if sem != nil {
					<-sem // Release slot
				}
				if r := recover(); r != nil {
					log.Printf("💥 Panic in handler for topic=%s: %v", kc.topic, r)
				}
			}()

			if err := handler(value); err != nil {
				log.Printf("❌ Handler error for topic=%s: %v", kc.topic, err)
			}
		}(msg.Value)
	}
}

func (kc *KafkaConsumer) Close() {
	if kc.reader == nil {
		return
	}
	if err := kc.reader.Close(); err != nil {
		log.Printf("⚠️ Error closing Kafka consumer for topic=%s: %v", kc.topic, err)
	} else {
		log.Printf("🧹 Kafka consumer closed for topic=%s", kc.topic)
	}
}
