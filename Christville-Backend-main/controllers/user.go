package controllers

import (
	"bossblock/db"
	"bossblock/model"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func GetOrCreateUser(c *gin.Context) {
	var input struct {
		TelegramID  string `json:"telegramId" binding:"required"`
		Username    string `json:"username"`
		ReferralKey string `json:"referralKey"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client := db.GetClient()
	usersCollection := client.Database("Christville").Collection("users")
	referralsCollection := client.Database("Christville").Collection("referrals")

	var user model.User
	err := usersCollection.FindOne(context.Background(), bson.M{"telegram_id": input.TelegramID}).Decode(&user)

	if err == mongo.ErrNoDocuments {
	
		newUser := model.User{
			TelegramID:     input.TelegramID,
			Username:       input.Username,
			TokenCount:     0,
			DailyVerseSeen: false, 
			QuizCompleted:  false, 
			QuizScore:      0,    
			LastLogin:      time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		referralKey, err := generateUniqueReferralKey(usersCollection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate referral key"})
			return
		}
		newUser.ReferralKey = referralKey

		var referrer model.User
		if input.ReferralKey != "" {
			err := usersCollection.FindOne(context.Background(), bson.M{"referral_key": input.ReferralKey}).Decode(&referrer)
			if err == nil {
				newUser.ReferredBy = referrer.ID.Hex()

				// Award referral bonus to the new user (adjust rewards as needed)
				newUser.TokenCount += 500 // Example reward for being referred

				// Award referral bonus to the referrer (adjust rewards as needed)
				referrerUpdate := bson.M{
					"$inc": bson.M{
						"token_count": 500, // Example reward for referring someone
					},
				}
				_, err = usersCollection.UpdateOne(context.Background(), bson.M{"telegram_id": referrer.TelegramID}, referrerUpdate)
				if err != nil {
					log.Printf("Failed to update referrer's bonus: %v", err)
				}
			}
		}

		_, err = usersCollection.InsertOne(context.Background(), newUser)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
            return
        }

        // Insert referral entry if applicable
        if newUser.ReferredBy != "" {
            referral := model.Referral{
                ReferrerID:     referrer.ID,
                ReferredID:     newUser.ID,
                DirectReferrer: true,
                CoinsEarned:    500, // Example reward for referring someone
                CreatedAt:      time.Now(),
            }

            if _, err = referralsCollection.InsertOne(context.Background(), referral); err != nil {
                log.Printf("Failed to create referral entry: %v", err)
            }
        }

        c.JSON(http.StatusCreated, gin.H{
            "user":    newUser,
            "isNew":   true,
        })
    } else if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
        return
    } else {
        c.JSON(http.StatusOK, gin.H{
            "user":   user,
            "isNew":  false,
        })
    }
}

func GetUserByID(c *gin.Context) {
	client := db.GetClient()
	database := client.Database("Christville") 


	userID := c.Param("userId")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user model.User

	err = database.Collection("users").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return the found user as JSON
	c.JSON(http.StatusOK, gin.H{"user": user})
}


func generateUniqueReferralKey(collection *mongo.Collection) (string, error) {
	for {
		key := model.GenerateUniqueHash()
		count, err := collection.CountDocuments(context.Background(), bson.M{"referral_key": key})
		if err != nil {
			return "", err
		}
		if count == 0 {
			return key, nil
		}
	}
}

func GetReferredUsers(c *gin.Context) {
	userID := c.Param("userId")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	databaseName := os.Getenv("Christville")
	client := db.GetClient()
	referralsCollection := client.Database(databaseName).Collection("referrals")
	usersCollection := client.Database(databaseName).Collection("users")
	referralEarningsCollection := client.Database(databaseName).Collection("referral_earnings")

	// Count total documents
	totalCount, err := referralsCollection.CountDocuments(context.Background(), bson.M{"referrer_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count referrals"})
		return
	}

	// Find referrals for the given userID with pagination
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize))
	cursor, err := referralsCollection.Find(
		context.Background(),
		bson.M{"referrer_id": objectID},
		findOptions,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusOK, gin.H{"referredUsers": []gin.H{}, "pagination": getPaginationInfo(page, pageSize, 0)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch referrals"})
		return
	}
	defer cursor.Close(context.Background())

	var referrals []model.Referral
	if err = cursor.All(context.Background(), &referrals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode referrals"})
		return
	}

	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	var response []gin.H
	var totalRecentEarnings int
	for _, referral := range referrals {
		var user model.User
		err := usersCollection.FindOne(context.Background(), bson.M{"_id": referral.ReferredID}).Decode(&user)
		if err != nil {
			continue // Skip this user if not found
		}
		//TODO: BATCHES FOR ALL DATA

		
		// Get recent earnings from referral_earnings
		var recentEarnings model.ReferralEarnings
		err = referralEarningsCollection.FindOne(
			context.Background(),
			bson.M{
				"referrer_id":    objectID,
				"referred_id":    referral.ReferredID,
				"last_earned_at": bson.M{"$gte": twentyFourHoursAgo},
			},
		).Decode(&recentEarnings)

		if err != nil && err != mongo.ErrNoDocuments {
			log.Printf("Error fetching recent earnings: %v", err)
		}

		totalRecentEarnings += recentEarnings.CoinsEarned

		response = append(response, gin.H{
			"id":             user.ID,
			"telegramId":     user.TelegramID,
			"username":       user.Username,
			"tokenCount":      user.TokenCount,
			"totalEarned":    referral.CoinsEarned,
			"recentEarnings": recentEarnings.CoinsEarned,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"referredUsers":       response,
		"pagination":          getPaginationInfo(page, pageSize, int(totalCount)),
		"totalRecentEarnings": totalRecentEarnings,
	})
}

func getPaginationInfo(page, pageSize, totalCount int) gin.H {
	totalPages := (totalCount + pageSize - 1) / pageSize
	return gin.H{
		"currentPage": page,
		"pageSize":    pageSize,
		"totalItems":  totalCount,
		"totalPages":  totalPages,
	}
}

func ClaimDailyBonus(c *gin.Context) {

    client := db.GetClient()
    database := client.Database("Christville")

    userID := c.Param("userId")
    log.Println("User ID:", userID)  // Debugging step

    // Convert user ID to ObjectID
    objectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    log.Println("Object ID:", objectID.Hex())  // Debugging step

    // Fetch user from database
    var user model.User
    err = database.Collection("users").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
    if err != nil {
        log.Printf("Failed to fetch user with ID %s: %v", objectID.Hex(), err)  // Debugging step
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    now := time.Now()
    today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    // Check if daily bonus has already been claimed today
    if !user.LastLogin.Before(today) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Daily bonus already claimed today"})
        return
    }

    // Set daily bonus amount (you can adjust this as needed)
    bonusTokens := 100 // Example: give 100 tokens as a daily bonus

    // Update user data with new token count and last login date
    userUpdate := bson.M{
        "$inc": bson.M{
            "token_count": bonusTokens,
        },
        "$set": bson.M{
            "last_login": today,
        },
    }

    _, err = database.Collection("users").UpdateOne(context.Background(), bson.M{"_id": objectID}, userUpdate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":    "Daily bonus claimed successfully",
        "bonusTokens": bonusTokens,
    })
}



func GetLeaderboard(c *gin.Context) {
	// databaseName := os.Getenv("Christville")
	client := db.GetClient()
	usersCollection := client.Database("Christville").Collection("users")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Get current user's ID
	currentUserID := c.Query("userId")

	// Calculate skip value for pagination
	skip := (page - 1) * pageSize

	// Get total number of users
	totalUsers, err := usersCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	// Get leaderboard
	findOptions := options.Find().
		SetSort(bson.D{{Key: "coin_count", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize))

	cursor, err := usersCollection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer cursor.Close(context.Background())

	var leaderboard []gin.H
	for cursor.Next(context.Background()) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		leaderboard = append(leaderboard, gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"tokenCount": user.TokenCount,
		})
	}

	// Get current user's rank
	var currentUserRank int64
	if currentUserID != "" {
		objectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			var currentUser model.User
			err = usersCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&currentUser)
			if err == nil {
				currentUserRank, _ = usersCollection.CountDocuments(
					context.Background(),
					bson.M{"token_count": bson.M{"$gt": currentUser.TokenCount}},
				)
				currentUserRank++ // Add 1 because ranks start at 1, not 0
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"pagination": gin.H{
			"currentPage": page,
			"pageSize":    pageSize,
			"totalItems":  totalUsers,
			"totalPages":  (totalUsers + int64(pageSize) - 1) / int64(pageSize),
		},
		"currentUserRank": currentUserRank,
	})
}
