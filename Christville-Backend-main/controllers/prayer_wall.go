package controllers

import (
	"context"
	"net/http"
	"time"
	"log"

	"bossblock/db"
	"bossblock/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

)

func CreatePrayer(c *gin.Context){
	client := db.GetClient()
	database := client.Database("Christville")

	var newPrayer struct {
		UserID string `json:"userId"`
		Prayer string `json:"prayer"`
	}

	if err := c.ShouldBindJSON(&newPrayer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(newPrayer.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	prayer := model.Prayer{
		ID:        primitive.NewObjectID(),
		UserID:    userObjectID,
		Prayer:    newPrayer.Prayer,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.Collection("prayers").InsertOne(ctx, prayer)
	if err != nil {
		log.Printf("Failed to insert prayer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add prayer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Prayer added successfully", "prayer_id": prayer.ID})
}


func GetAllPrayers(c *gin.Context) {
	client := db.GetClient()
	database := client.Database("Christville") 

	var prayers []model.Prayer

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Query all prayers from MongoDB
	cursor, err := database.Collection("prayers").Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to fetch prayers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prayers"})
		return
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor and decode each prayer
	for cursor.Next(ctx) {
		var prayer model.Prayer
		if err := cursor.Decode(&prayer); err != nil {
			log.Printf("Failed to decode prayer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode prayer"})
			return
		}
		prayers = append(prayers, prayer)
	}

	// Handle any errors encountered during iteration
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	// Return all prayers as JSON response
	c.JSON(http.StatusOK, gin.H{"prayers": prayers})
}

func GetUserPrayers(c *gin.Context) {
    client := db.GetClient()
    database := client.Database("Christville") 

    userID := c.Param("userId")
    userObjectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var prayers []model.Prayer

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := database.Collection("prayers").Find(ctx, bson.M{"user_id": userObjectID})
    if err != nil {
        log.Printf("Failed to fetch user's prayers: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user's prayers"})
        return
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var prayer model.Prayer
        if err := cursor.Decode(&prayer); err != nil {
            log.Printf("Failed to decode prayer: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode prayer"})
            return
        }
        prayers = append(prayers, prayer)
    }

    if err := cursor.Err(); err != nil {
        log.Printf("Cursor error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"user_prayers": prayers})
}