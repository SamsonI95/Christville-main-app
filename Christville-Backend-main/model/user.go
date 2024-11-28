package model

import (
	"bossblock/db"
	"context"
	"crypto/rand"
	"encoding/hex"
	//"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TelegramID       string             `bson:"telegram_id" json:"telegramId"`
	Username         string             `bson:"username,omitempty" json:"username,omitempty"`
	TokenCount       int                `bson:"token_count,omitempty" json:"tokenCount,omitempty"` // Tokens earned from activities
	DailyVerseSeen   bool               `bson:"daily_verse_seen,omitempty" json:"dailyVerseSeen,omitempty"` // Tracks if the user has seen the daily verse
	QuizCompleted    bool               `bson:"quiz_completed,omitempty" json:"quizCompleted,omitempty"` // Tracks if the user has completed the daily quiz
	QuizScore        int                `bson:"quiz_score,omitempty" json:"quizScore,omitempty"` // Stores the user's score from the quiz
	LastLogin        time.Time          `bson:"last_login,omitempty" json:"lastLogin,omitempty"` // Last login timestamp
	
	// New fields for referral system
	ReferralKey      string             `bson:"referral_key,omitempty" json:"referralKey,omitempty"` // Unique referral key for this user
	ReferredBy       string             `bson:"referred_by,omitempty" json:"referredBy,omitempty"` // The referral key of the person who referred this user
	
	CreatedAt        time.Time          `bson:"created_at,omitempty" json:"createdAt,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at,omitempty" json:"updatedAt,omitempty"`
}



func GenerateUniqueHash() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func CreateUser(user *User) error {
	collection := db.GetClient().Database("Christville").Collection("users")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), user)
	return err
}

func GetUserByTelegramID(telegramID string) (*User, error) {
	collection := db.GetClient().Database("Christville").Collection("users")
	var user User
	err := collection.FindOne(context.Background(), bson.M{"telegram_id": telegramID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user *User) error {
	collection := db.GetClient().Database("Christville").Collection("users")
	user.UpdatedAt = time.Now()
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

func GenerateUniqueReferralKey(client *mongo.Client) (string, error) {
	collection := client.Database("Christville").Collection("users")
	for {
		key := GenerateUniqueHash()
		count, err := collection.CountDocuments(context.Background(), bson.M{"referral_key": key})
		if err != nil {
			return "", err
		}
		if count == 0 {
			return key, nil
		}
	}
}
