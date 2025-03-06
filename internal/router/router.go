package router

import (
	"ikurotime/backlog-go-backend/internal/projects"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewServer(client *mongo.Client) *gin.Engine {
	r := gin.Default()
	ProjectRouter(r, client)
	return r
}

func ProjectRouter(r *gin.Engine, client *mongo.Client) {
	router := r.Group("/projects")
	projectHandler := projects.NewRouteHandler(client)
	router.GET("", projectHandler.GetAll)
}
