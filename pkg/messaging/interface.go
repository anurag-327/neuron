package messaging

import (
	"context"
)

type Subscriber interface {
	Consume(ctx context.Context, handler func(message []byte) error)
	Close()
}

type Publisher interface {
	Publish(topic string, data []byte) error
	Close()
}
