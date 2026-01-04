package conn

import (
	"context"
	"log"
	"os"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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

	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		log.Fatal("Error retrieving default mongo config:", err)
	}

	// Verify connection with a Ping
	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal("Could not ping MongoDB:", err)
	}

	log.Println("Connected to MongoDB!")
}
