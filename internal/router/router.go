package router

import (
	"ikurotime/backlog-go-backend/internal/ideas"
	"ikurotime/backlog-go-backend/internal/projects"
	"log"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Router struct {
	engine *gin.Engine
	client *mongo.Client
}

func NewRouter(client *mongo.Client) *Router {
	r := &Router{
		engine: gin.Default(),
		client: client,
	}

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	r.setupPublicRoutes()
	r.setupProtectedRoutes()
}

func (r *Router) setupPublicRoutes() {
	r.engine.GET("/health", r.handleHealth())
}

func (r *Router) setupProtectedRoutes() {
	api := r.engine.Group("/api")
	{
		// Projects routes
		projectsGroup := api.Group("/projects")
		{
			handler := projects.NewHandler(r.client)
			projectsGroup.GET("", handler.GetAll) // Public route
		}

		// Ideas routes
		ideasGroup := api.Group("/ideas")
		{
			handler := ideas.NewHandler(r.client)
			ideasGroup.GET("", handler.GetAll) // Public route
			ideasGroup.POST("/:id/like", r.requireAuth(), handler.LikeIdea)
			ideasGroup.DELETE("/:id/like", r.requireAuth(), handler.UnlikeIdea)
		}
	}
}

// requireAuth is a middleware that checks for authentication
func (r *Router) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if sessionToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Missing authentication token",
			})
			return
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			log.Printf("JWT verification failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Invalid authentication token",
			})
			return
		}

		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			log.Printf("Failed to get user information: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Failed to get user information",
			})
			return
		}

		c.Set("user_id", usr.ID)
		c.Set("user_banned", usr.Banned)
		c.Set("user_email", usr.EmailAddresses[0].EmailAddress)

		c.Next()
	}
}

func (r *Router) handleHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "backlog-go-backend",
		})
	}
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
