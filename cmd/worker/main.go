package main

import (
	"log"

	"github.com/anurag-327/neuron/db"
	"github.com/anurag-327/neuron/internal/consumer"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db.ConnectMongoDB()
}

func main() {
	topics := []string{}
	topics = append(topics, "code-jobs")

	c, _ := consumer.NewConsumer("code-runner-group", topics)

	c.Close()

}
