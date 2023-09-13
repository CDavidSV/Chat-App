package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var mongoOnce sync.Once

func MongoClient() *mongo.Client {
	mongoOnce.Do(func() {
		// Set client options
		MONGOURI := os.Getenv("MONGO_URI")
		clientOptions := options.Client().ApplyURI(MONGOURI)

		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			panic(err)
		}

		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			panic(err)
		}

		clientInstance = client
		fmt.Println("Connected to MongoDB")
	})

	return clientInstance
}
