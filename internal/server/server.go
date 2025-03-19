package server

import (
	"ikurotime/backlog-go-backend/config"
	"ikurotime/backlog-go-backend/internal/router"

	"github.com/clerk/clerk-sdk-go/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Server struct {
	router *router.Router
	client *mongo.Client
}

func NewServer(client *mongo.Client) (*Server, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	clerk.SetKey(cfg.ClerkConfig.ApiKey)

	s := &Server{
		router: router.NewRouter(client),
		client: client,
	}

	return s, nil
}

func (s *Server) Run(addr string) error {
	return s.router.GetEngine().Run(addr)
}
