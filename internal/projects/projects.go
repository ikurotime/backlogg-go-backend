package projects

import (
	"context"
	"ikurotime/backlog-go-backend/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/*
	Projects

- GetAll
- GetByID
- Creat
- Update
- Delete
- GetByUserID
- SaveFavorite
- RemoveFavorite
- GetFavorites
*/
type Project struct {
	ID          string `bson:"id"`
	Title       string `bson:"title"`
	Description string `bson:"description"`
}

type ProjectDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type RouteHandler struct {
	client *mongo.Client
}

func NewRouteHandler(client *mongo.Client) *RouteHandler {
	return &RouteHandler{
		client: client,
	}
}

func (h *RouteHandler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load config: " + err.Error(),
		})
		return
	}

	collection := h.client.Database(cfg.MongoDBConfig.Database).Collection("projects")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch projects: " + err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var projects []Project
	if err := cursor.All(ctx, &projects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode projects",
		})
		return
	}

	projectsDTO := make([]ProjectDTO, 0, len(projects))
	for _, project := range projects {
		projectsDTO = append(projectsDTO, ProjectDTO(project))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": projectsDTO,
	})
}
