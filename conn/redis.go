package conn

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func GetRedisClient(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		Protocol: 2,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis")
	}
	log.Printf("Redis Connected")
	return client, nil
}

func ConnectRedisDb() error {
	ctx := context.Background()
	client, err := GetRedisClient(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	RedisClient = client
	return nil
}
