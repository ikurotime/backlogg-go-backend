package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ikurotime/backlog-go-backend/internal/router"
	"ikurotime/backlog-go-backend/pkg/mongodbx"
)

func main() {
	mongoClient := mongodbx.ConnectMongoDB()

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	server := router.NewServer(mongoClient)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(":8080"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
}
