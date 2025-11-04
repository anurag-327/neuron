package db

import (
	"log"
	"os"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGO_URI OR MONGO_DB_NAME environment variable is not set")
	}
	err := mgm.SetDefaultConfig(nil, dbName, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
}
