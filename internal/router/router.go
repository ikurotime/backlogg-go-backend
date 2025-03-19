package router

import (
	"ikurotime/backlog-go-backend/config"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	// Setup CORS middleware
	r.engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.Server.AllowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

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
	api := r.engine.Group("/v1")
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
			ideasGroup.POST("/:id/bookmark", r.requireAuth(), handler.BookmarkIdea)
			ideasGroup.DELETE("/:id/bookmark", r.requireAuth(), handler.UnbookmarkIdea)
			ideasGroup.GET("/bookmarks", r.requireAuth(), handler.GetBookmarkedIdeas)
		}
	}
}

// requireAuth is a middleware that checks for authentication
func (r *Router) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionToken string

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// If no token in header, try to get it from cookie
		if sessionToken == "" {
			cookie, err := c.Cookie("__session")
			if err == nil && cookie != "" {
				sessionToken = cookie
			}
		}

		// If still no token, authentication fails
		if sessionToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Missing authentication token",
			})
			return
		}

		// Verify the token
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

		// Get user details
		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			log.Printf("Failed to get user information: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Failed to get user information",
			})
			return
		}

		// Store user information in the context
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
