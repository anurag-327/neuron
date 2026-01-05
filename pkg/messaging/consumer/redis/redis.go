package redisConsumer

import (
	"context"
	"log"
	"time"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/pkg/messaging"
	"github.com/redis/go-redis/v9"
)

type RedisConsumer struct {
	client *redis.Client
	stream string
	group  string
}

func NewConsumer(group, stream string) (messaging.Subscriber, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := conn.GetRedisClient(ctx)
	if err != nil {
		log.Printf("Redis connection failed: %v\n", err)
		return nil, err
	}

	// Create stream group if it does not exist
	err = client.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed creating Redis group: %v\n", err)
		return nil, err
	}

	log.Printf("Redis consumer initialized. Group=%s Stream=%s", group, stream)
	return &RedisConsumer{client: client, stream: stream, group: group}, nil
}

func (rc *RedisConsumer) Consume(ctx context.Context, handler func([]byte) error) {
	rc.ConsumeControlled(ctx, handler, 0)
}

func (rc *RedisConsumer) ConsumeControlled(ctx context.Context, handler func([]byte) error, maxConcurrent int) {
	var sem chan struct{}
	if maxConcurrent > 0 {
		sem = make(chan struct{}, maxConcurrent)
		log.Printf("Controlled Redis consumer started for stream=%s (limit=%d)", rc.stream, maxConcurrent)
	} else {
		log.Printf("Unbounded Redis consumer started for stream=%s", rc.stream)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Redis consumer context canceled (stream=%s)", rc.stream)
			return
		default:
		}

		if sem != nil {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}

		msgs, err := rc.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    rc.group,
			Consumer: "worker-1",
			Streams:  []string{rc.stream, ">"},
			Count:    1,
			Block:    0, // block indefinitely
		}).Result()

		if err != nil {
			if sem != nil {
				<-sem
			}

			if err == context.Canceled {
				return
			}

			log.Printf("Redis stream read error [%s]: %v", rc.stream, err)

			// Simple backoff
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range msgs {
			for _, message := range stream.Messages {

				value := []byte(message.Values["value"].(string))
				msgID := message.ID

				// Process concurrently
				go func(msgID string, payload []byte) {
					defer func() {
						if sem != nil {
							<-sem
						}
						if r := recover(); r != nil {
							log.Printf("Panic in handler for stream=%s: %v", rc.stream, r)
						}
					}()

					if err := handler(payload); err != nil {
						log.Printf("Handler error for stream=%s: %v", rc.stream, err)
						return
					}

					// ACK message
					if err := rc.client.XAck(ctx, rc.stream, rc.group, msgID).Err(); err != nil {
						log.Printf("Failed to ACK Redis msg=%s err=%v", msgID, err)
					}
					if err := rc.client.XDel(ctx, rc.stream, msgID).Err(); err != nil {
						log.Printf("Failed to Delete Job From Queue=%s err=%v", msgID, err)
					}

				}(msgID, value)
			}
		}
	}
}

func (rc *RedisConsumer) Close() {
	if rc.client != nil {
		if err := rc.client.Close(); err != nil {
			log.Printf("Redis client closed successfully")
		} else {
			log.Printf("Failed to close redis client")
		}
	}
}

func (rc *RedisConsumer) Health() error {
	return rc.client.Ping(context.Background()).Err()
}
