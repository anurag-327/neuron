package messaging

import (
	"context"
)

type Subscriber interface {
	Consume(ctx context.Context, handler func(message []byte) error)
	ConsumeControlled(ctx context.Context, handler func(message []byte) error, maxConcurrent int)
	Close()
}

type Publisher interface {
	Publish(topic string, key string, data []byte) error
	Close()
}
