package main

import (
	"log"

	"github.com/anurag-327/neuron/db"
	"github.com/anurag-327/neuron/internal/factory"
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
	topic := "code-jobs"

	c := factory.GetSubscriber("code-runner-group", topic)

	c.Close()

}
