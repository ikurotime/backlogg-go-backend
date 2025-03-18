package router

import (
	"fmt"
	"ikurotime/backlog-go-backend/config"
	"ikurotime/backlog-go-backend/internal/projects"
	"log"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func clerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Invalid or missing authentication token",
			})
			return
		}

		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Failed to get user information",
			})
			return
		}

		// Store user information in the context for later use
		c.Set("user_id", usr.ID)
		c.Set("user_banned", usr.Banned)

		c.Next()
	}
}

func setupPublicRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "backlog-go-backend",
		})
	})
}

func setupProtectedRoutes(r *gin.Engine, client *mongo.Client) {
	api := r.Group("/api")
	api.Use(clerkAuthMiddleware())
	{
		projectsGroup := api.Group("/projects")
		{
			handler := projects.NewHandler(client)
			projectsGroup.GET("", handler.GetAll)

		}
	}
}

func NewServer(client *mongo.Client) *gin.Engine {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	clerk.SetKey(cfg.ClerkConfig.ApiKey)

	r := gin.Default()

	setupPublicRoutes(r)
	setupProtectedRoutes(r, client)

	return r
}
