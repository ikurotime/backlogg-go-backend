package ideas

import (
	"context"
	"fmt"
	"ikurotime/backlog-go-backend/config"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Idea represents a project idea in the database
type Idea struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string        `bson:"title" json:"title"`
	Description   string        `bson:"description" json:"description"`
	Tags          []string      `bson:"tags" json:"tags"`
	Difficulty    string        `bson:"difficulty" json:"difficulty"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
	AuthorID      string        `bson:"author_id" json:"author_id"`
	LikesCount    int           `bson:"likes_count" json:"likes_count"`
	CommentsCount int           `bson:"comments_count" json:"comments_count"`
	Details       interface{}   `bson:"details,omitempty" json:"details,omitempty"`
}

// Like represents a user's like on an idea
type Like struct {
	ID        string    `bson:"_id,omitempty"`
	UserID    string    `bson:"user_id"`
	IdeaID    string    `bson:"idea_id"`
	CreatedAt time.Time `bson:"created_at"`
}

// Comment represents a user's comment on an idea
type Comment struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	IdeaID    string    `bson:"idea_id" json:"idea_id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Content   string    `bson:"content" json:"content"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Bookmark represents a user's bookmark on an idea
type Bookmark struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	IdeaID    bson.ObjectID `bson:"idea_id"`
	CreatedAt time.Time     `bson:"created_at"`
}

// Handler handles idea-related HTTP requests
type Handler struct {
	client *mongo.Client
}

// NewHandler creates a new ideas handler
func NewHandler(client *mongo.Client) *Handler {
	return &Handler{
		client: client,
	}
}

// setupIndexes creates necessary indexes for optimal query performance
func (h *Handler) setupIndexes(ctx context.Context, db *mongo.Database) error {
	// Ideas collection indexes
	ideasColl := db.Collection("ideas")
	_, err := ideasColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tags", Value: 1},
				{Key: "difficulty", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "likes_count", Value: -1},
				{Key: "comments_count", Value: -1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "author_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	})
	if err != nil {
		return err
	}

	// Likes collection indexes
	likesColl := db.Collection("likes")
	_, err = likesColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "idea_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "idea_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	})
	if err != nil {
		return err
	}

	// Comments collection indexes
	commentsColl := db.Collection("comments")
	_, err = commentsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "idea_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	})
	return err
}

// GetAll retrieves ideas with optional filtering and sorting
func (h *Handler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	db := h.client.Database(cfg.MongoDBConfig.Database)
	collection := db.Collection("ideas")

	// Build filter based on query parameters
	filter := bson.M{}
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		filter["difficulty"] = difficulty
	}

	// Build sort options
	sort := bson.D{{Key: "created_at", Value: -1}}
	if sortBy := c.Query("sort"); sortBy != "" {
		switch sortBy {
		case "trending":
			sort = bson.D{
				{Key: "likes_count", Value: -1},
				{Key: "comments_count", Value: -1},
				{Key: "created_at", Value: -1},
			}
		case "popular":
			sort = bson.D{{Key: "likes_count", Value: -1}}
		}
	}

	// Execute query with pagination
	page := 1
	pageSize := 20 // Default page size

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if size := c.Query("size"); size != "" {
		if parsedSize, err := strconv.Atoi(size); err == nil && parsedSize > 0 && parsedSize <= 100 {
			pageSize = parsedSize
		}
	}

	skip := int64((page - 1) * pageSize)

	// Get total count for pagination
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count ideas"})
		return
	}

	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(pageSize))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ideas"})
		return
	}
	defer cursor.Close(ctx)

	var ideas []Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode ideas", "message": err.Error()})
		return
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"data": ideas,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}

func (h *Handler) GetOne(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}
	db := h.client.Database(cfg.MongoDBConfig.Database)
	collection := db.Collection("ideas")
	ideaID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid idea ID"})
		return
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: ideaID}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "idea_details"},
			{Key: "let", Value: bson.D{{Key: "ideaId", Value: "$_id"}}},
			{Key: "pipeline", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{"$idea_id", "$$ideaId"}}}}}}},
				bson.D{{Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "idea_id", Value: 0},
				}}},
			}},
			{Key: "as", Value: "details"},
		}}},
		{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$details"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch idea", "message": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var results []Idea
	if err = cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode idea", "message": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idea not found"})
		return
	}

	idea := results[0]
	fmt.Printf("Fetched idea with details: %+v\n", idea)

	c.JSON(http.StatusOK, gin.H{
		"data": idea,
	})
}

// LikeIdea handles liking an idea with transaction support
func (h *Handler) LikeIdea(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetString("user_id")
	ideaID := c.Param("id")

	db := h.client.Database(cfg.MongoDBConfig.Database)

	// Start transaction
	session, err := db.Client().StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		// Check if like already exists
		likesColl := db.Collection("likes")
		exists, err := likesColl.CountDocuments(sessCtx, bson.M{
			"user_id": userID,
			"idea_id": ideaID,
		})
		if err != nil {
			return nil, err
		}
		if exists > 0 {
			return nil, nil // Like already exists
		}

		// Insert like
		_, err = likesColl.InsertOne(sessCtx, Like{
			UserID:    userID,
			IdeaID:    ideaID,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return nil, err
		}

		// Increment likes_count in ideas collection
		ideasColl := db.Collection("ideas")
		_, err = ideasColl.UpdateOne(
			sessCtx,
			bson.M{"_id": ideaID},
			bson.M{"$inc": bson.M{"likes_count": 1}},
		)
		return nil, err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like idea"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idea liked successfully"})
}

// UnlikeIdea handles unliking an idea with transaction support
func (h *Handler) UnlikeIdea(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	userID := c.GetString("user_id")
	ideaID := c.Param("id")

	db := h.client.Database(cfg.MongoDBConfig.Database)

	session, err := db.Client().StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		// Delete like
		likesColl := db.Collection("likes")
		result, err := likesColl.DeleteOne(sessCtx, bson.M{
			"user_id": userID,
			"idea_id": ideaID,
		})
		if err != nil {
			return nil, err
		}
		if result.DeletedCount == 0 {
			return nil, nil // Like didn't exist
		}

		// Decrement likes_count in ideas collection
		ideasColl := db.Collection("ideas")
		_, err = ideasColl.UpdateOne(
			sessCtx,
			bson.M{"_id": ideaID},
			bson.M{"$inc": bson.M{"likes_count": -1}},
		)
		return nil, err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike idea"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idea unliked successfully"})
}

// BookmarkIdea handles bookmarking an idea
func (h *Handler) BookmarkIdea(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	userID := c.GetString("user_id")
	ideaID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid idea ID"})
		return
	}

	db := h.client.Database(cfg.MongoDBConfig.Database)
	bookmarksColl := db.Collection("bookmarks")

	exists, err := bookmarksColl.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"idea_id": ideaID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bookmark"})
		return
	}
	if exists > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Already bookmarked"})
		return
	}

	_, err = bookmarksColl.InsertOne(ctx, Bookmark{
		UserID:    userID,
		IdeaID:    ideaID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bookmark idea"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idea bookmarked successfully"})
}

// UnbookmarkIdea handles unbookmarking an idea
func (h *Handler) UnbookmarkIdea(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	userID := c.GetString("user_id")
	ideaID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid idea ID"})
		return
	}

	db := h.client.Database(cfg.MongoDBConfig.Database)
	bookmarksColl := db.Collection("bookmarks")

	result, err := bookmarksColl.DeleteOne(ctx, bson.M{
		"user_id": userID,
		"idea_id": ideaID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unbookmark idea"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Bookmark not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idea unbookmarked successfully"})
}

// GetBookmarkedIdeas retrieves bookmarked ideas for a user
func (h *Handler) GetBookmarkedIdeas(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	userID := c.GetString("user_id")
	db := h.client.Database(cfg.MongoDBConfig.Database)

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if size := c.Query("size"); size != "" {
		if parsedSize, err := strconv.Atoi(size); err == nil && parsedSize > 0 && parsedSize <= 100 {
			pageSize = parsedSize
		}
	}

	skip := int64((page - 1) * pageSize)

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$sort", Value: bson.M{"created_at": -1}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: pageSize}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "ideas",
			"localField":   "idea_id",
			"foreignField": "_id",
			"as":           "idea",
		}}},
		{{Key: "$unwind", Value: "$idea"}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$idea"}}},
	}

	cursor, err := db.Collection("bookmarks").Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarked ideas"})
		return
	}
	defer cursor.Close(ctx)

	var ideas []Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode ideas"})
		return
	}

	total, err := db.Collection("bookmarks").CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count bookmarks"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"data": ideas,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}
