package redisLogger

// Config holds all configuration for the logger.
type Config struct {
	ServiceName string
	RedisAddr   string
	RedisUser   string
	RedisPass   string
	RedisDB     int
	QueueName   string
}
