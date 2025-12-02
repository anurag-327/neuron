package factory

import (
	"log"
	"sync"

	"github.com/anurag-327/neuron/pkg/messaging"
	kafkaConsumer "github.com/anurag-327/neuron/pkg/messaging/consumer/kafka"
	kafkaProducer "github.com/anurag-327/neuron/pkg/messaging/producer/kafka"
)

var (
	publisherInstance  messaging.Publisher
	subscriberInstance messaging.Subscriber
	oncePublisher      sync.Once
	onceSubscriber     sync.Once
)

func GetPublisher() messaging.Publisher {
	oncePublisher.Do(func() {
		p, err := kafkaProducer.NewProducer()
		if err != nil {
			log.Fatalf("Failed to init Kafka publisher: %v", err)
		}
		publisherInstance = p
	})
	return publisherInstance
}

func GetSubscriber(group string, topic string) messaging.Subscriber {
	onceSubscriber.Do(func() {
		s, err := kafkaConsumer.NewConsumer(group, topic)
		if err != nil {
			log.Fatalf("Failed to init Kafka subscriber: %v", err)
		}
		subscriberInstance = s
	})
	return subscriberInstance
}
