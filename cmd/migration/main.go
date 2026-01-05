package migration

import (
	"log"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn.ConnectMongoDB()
}

func main() {
	models.CreateUserIndexes()
	models.CreateCreditTransactionIndexes()
	models.CreateJobIndexes()
	models.CreateApiLogIndexes()
	models.CreateSystemStatusIndexes()
}
