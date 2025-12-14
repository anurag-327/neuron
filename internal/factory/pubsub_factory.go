package factory

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/anurag-327/neuron/pkg/messaging"
	kafkaConsumer "github.com/anurag-327/neuron/pkg/messaging/consumer/kafka"
	redisConsumer "github.com/anurag-327/neuron/pkg/messaging/consumer/redis"
	kafkaProducer "github.com/anurag-327/neuron/pkg/messaging/producer/kafka"
	redisProducer "github.com/anurag-327/neuron/pkg/messaging/producer/redis"
)

var (
	publisherInstance messaging.Publisher
	publisherErr      error
	oncePublisher     sync.Once
)

func GetPublisher() (messaging.Publisher, error) {
	oncePublisher.Do(func() {
		backend := os.Getenv("QUEUE_SERVICE")

		switch backend {
		case "kafka":
			publisherInstance, publisherErr = kafkaProducer.NewProducer()
		case "redis", "":
			publisherInstance, publisherErr = redisProducer.NewProducer()
		default:
			publisherErr = fmt.Errorf("unsupported QUEUE_BACKEND: %s", backend)
		}
	})

	if publisherErr != nil {
		return nil, publisherErr
	}

	return publisherInstance, nil
}

var (
	consumerRegistry = make(map[string]messaging.Subscriber)
	consumerErrors   = make(map[string]error)
	consumerMu       sync.Mutex
)

func GetSubscriber(group string, topic string) (messaging.Subscriber, error) {
	key := group + ":" + topic

	consumerMu.Lock()
	defer consumerMu.Unlock()

	// If already created → return it (even if error)
	if sub, exists := consumerRegistry[key]; exists {
		return sub, consumerErrors[key]
	}

	backend := os.Getenv("QUEUE_SERVICE")

	var (
		s   messaging.Subscriber
		err error
	)

	switch backend {
	case "kafka":
		s, err = kafkaConsumer.NewConsumer(group, topic)
	case "redis", "":
		s, err = redisConsumer.NewConsumer(group, topic)
	default:
		err = fmt.Errorf("unsupported QUEUE_BACKEND: %s", backend)
	}

	// store result (subscriber or nil)
	consumerRegistry[key] = s
	consumerErrors[key] = err

	// caller handles error
	return s, err
}

func StartConsumer(ctx context.Context, topic string, group string, maxConcurrent int, handler func(jobBytes []byte) error) error {
	sub, err := GetSubscriber(group, topic)
	if err != nil {
		return err
	}

	go func(sub messaging.Subscriber) {
		defer sub.Close()
		sub.ConsumeControlled(ctx, handler, maxConcurrent)
	}(sub)

	log.Printf("Worker listening on topic: %s", topic)
	return nil
}
