package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"bossblock/db"
	"bossblock/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FetchRandomBibleVerse() (model.BibleVerse, error) {
	url := "https://bible-api.com/?random=verse"

	resp, err := http.Get(url)
	if err != nil {
		return model.BibleVerse{}, fmt.Errorf("failed to fetch Bible verse: %v", err)
	}
	defer resp.Body.Close()

	var apiResponse struct {
		Reference string `json:"reference"` // Extracting only "reference" field
		Text      string `json:"text"`      // Extracting only "text" field
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return model.BibleVerse{}, fmt.Errorf("failed to decode API response: %v", err)
	}

	bibleVerse := model.BibleVerse{
		Reference: apiResponse.Reference,
		Text:      apiResponse.Text,
	}

	
	client := db.GetClient() // Get the MongoDB client instance
	database := client.Database("Christville") // Replace with your actual database name

	update := bson.M{
		"$set": bson.M{
			"reference": bibleVerse.Reference,
			"text":      bibleVerse.Text,
			"timestamp": time.Now(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.Collection("daily_verses").UpdateOne(
		ctx,
		bson.M{}, // Empty filter means we will replace the only document in this collection
		update,
		options.Update().SetUpsert(true), // Upsert ensures we create a new document if none exists
	)

	if err != nil {
        return model.BibleVerse{}, fmt.Errorf("failed to store Bible verse in MongoDB: %v", err)
    }

	return bibleVerse, nil
}


func GetDailyBibleVerse(c *gin.Context) {
	client := db.GetClient()
	db := client.Database("Christville")

	var cachedVerse model.CachedVerse

	err := db.Collection("daily_verses").FindOne(c.Request.Context(), bson.M{}).Decode(&cachedVerse)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "No daily Bible verse found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve daily Bible verse: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reference": cachedVerse.Reference,
        "text":      cachedVerse.Text,
        "timestamp": cachedVerse.Timestamp,
    })
}
