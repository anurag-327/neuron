package redisProducer

import (
	"context"
	"log"
	"time"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/pkg/messaging"
	"github.com/redis/go-redis/v9"
)

type RedisProducer struct {
	client *redis.Client
}

func NewProducer() (messaging.Publisher, error) {
	ctx := context.Background()
	client, err := conn.GetRedisClient(ctx)
	if err != nil {
		log.Printf(" Redis connection failed: %v\n", err)
		return nil, err
	}
	return &RedisProducer{client: client}, nil
}

func (rp *RedisProducer) Publish(stream string, key string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msgID, err := rp.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"key":   key,
			"value": data,
		},
	}).Result()

	if err != nil {
		log.Printf("Failed to publish to Redis stream '%s': %v", stream, err)
		return err
	}

	log.Printf("Redis delivered message to stream '%s' | ID=%s", stream, msgID)

	return nil
}

func (rp *RedisProducer) Close() {
	if rp.client != nil {
		if err := rp.client.Close(); err != nil {
			log.Printf("Redis client closed successfully")
		} else {
			log.Printf("Failed to close redis client")
		}
	}

}
