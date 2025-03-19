package main

import (
	"context"
	"log"

	"ikurotime/backlog-go-backend/internal/server"
	"ikurotime/backlog-go-backend/pkg/mongodbx"
)

func main() {
	// Initialize MongoDB client
	client := mongodbx.ConnectMongoDB()
	defer client.Disconnect(context.Background())

	// Create and start server
	server, err := server.NewServer(client)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
