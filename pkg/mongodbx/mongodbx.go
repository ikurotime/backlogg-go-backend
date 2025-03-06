package mongodbx

import (
	"context"
	"ikurotime/backlog-go-backend/config"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB() *mongo.Client {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	uri := cfg.MongoDBConfig.Protocol + "://" + cfg.MongoDBConfig.User + ":" + cfg.MongoDBConfig.Password + "@" + cfg.MongoDBConfig.Host + ":" + cfg.MongoDBConfig.Port + "/?directConnection=true"
	log.Print(uri)

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal(err)
	}

	log.Print("[ --- Connected to MongoDB --- ]")
	return client
}
